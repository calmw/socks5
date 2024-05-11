package protocol

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/calmw/socks5/logger"
	"io"
	"log"
	"net"
	"strings"
)

const (
	Socks5Version         = uint8(5)
	AUTH                  = 1 // 代理服务器是否需要账号密码认证, 1需要认证，0无需认证
	ErrNetworkUnreachable = 2
	ErrHostUnreachable    = 3
	ErrConnectionRefused  = 4
)

type Socks5 struct {
	AuthMethod uint8 // 验证方式
}

func NewSocks5() *Socks5 {
	return &Socks5{}
}

func (s *Socks5) Auth(reader *bufio.Reader, conn net.Conn) (*AuthContext, error) {
	// 认证方法
	methods, err := readMethods(reader)
	if err != nil {
		logger.Zap.Sugar().Info(err)
		return nil, err
	}

	var method uint8 = AuthMethodNo
	var methodCheck bool // 必须有无需身份认证否则返回没可接受的方法
	var isClientSupportAuth bool
	for _, m := range methods {
		if m == AuthMethodNo {
			methodCheck = true
		}
		if m == AuthMethodUsernamePwd {
			isClientSupportAuth = true
			break
		}
	}
	// 服务端要求认证
	// 客户端支持账号密码的时候方法会有0，2，不支持的话只会有0
	if isClientSupportAuth && AUTH > 0 {
		method = AuthMethodUsernamePwd
	}
	if !methodCheck {
		method = AuthMethodUnSupport
	}

	s.AuthMethod = method

	_, err = conn.Write([]byte{Socks5Version, method})
	if err != nil {
		logger.Zap.Sugar().Info(err)
		return nil, err
	}

	if method == AuthMethodUsernamePwd {
		var authUsernamePasswordVersion uint8
		var username, pwd string
		status := uint8(0)

		// 如果获取用户名密码失败，也算认证失败
		authUsernamePasswordVersion, username, pwd, err = s.parseAuthInfo(reader, conn)
		if err == nil {
			err = s.checkAuthInfo(username, pwd)
			if err != nil {
				status = uint8(1) // 大于0认证失败
			}
		}

		// 注意，这里使用authUsernamePasswordVersion，而不是socks版本5
		_, err = conn.Write([]byte{authUsernamePasswordVersion, status})

		return &AuthContext{
			Method: AuthMethodUsernamePwd,
			Payload: map[string]string{
				"username": username,
				"pwd":      pwd,
			},
		}, nil
	}

	return nil, nil
}

func (s *Socks5) parseAuthInfo(reader *bufio.Reader, conn net.Conn) (uint8, string, string, error) {
	authUsernamePasswordVersion := uint8(0)
	// 获取用户名
	readerL := io.LimitReader(reader, 2)
	var buf [2]byte
	_, err := readerL.Read(buf[:])
	if err != nil {
		logger.Zap.Sugar().Info(err)
		return 0, "", "", err
	}
	authUsernamePasswordVersion = buf[0]
	usernameLen := int(buf[1])
	readerL = io.LimitReader(reader, int64(buf[1]))
	usernameBytes := make([]byte, usernameLen)
	_, err = readerL.Read(usernameBytes)
	if err != nil {
		logger.Zap.Sugar().Info(err)
		return 0, "", "", err
	}
	username := string(usernameBytes)

	// 获取密码
	var pwdLenBuf [1]byte
	readerL = io.LimitReader(reader, 1)
	_, err = readerL.Read(pwdLenBuf[:])
	if err != nil {
		logger.Zap.Sugar().Info(err)
		return 0, "", "", err
	}
	pwdBytes := make([]byte, int(pwdLenBuf[0]))
	readerL = io.LimitReader(reader, int64(pwdLenBuf[0]))
	_, err = readerL.Read(pwdBytes)
	if err != nil {
		logger.Zap.Sugar().Info(err)
		return 0, "", "", err
	}
	pwd := string(pwdBytes)

	return authUsernamePasswordVersion, username, pwd, err
}

func (s *Socks5) checkAuthInfo(username, pwd string) error {
	log.Printf("用户名:%s,密码:%s\n", username, pwd)
	if username != "cisco" || pwd != "123456" {
		log.Println("用户验证失败")
		return errors.New("username or password error")
	}
	log.Println("用户验证通过")
	return nil
}

func (s *Socks5) CreateProxy(reader io.Reader, conn net.Conn) (net.Conn, error) {
	var buf [4]byte
	_, err := reader.Read(buf[:])
	if err != nil {
		logger.Zap.Sugar().Info(err)
		return nil, err
	}

	// 说明：这里的DST.ADDR和DST.PORT在COMMAND不同时有不用的表示
	// CONNECT 希望连接的target服务器ip地址和端口号
	// BIND 希望连接的target服务器ip地址和端口号
	// UDP ASSOCIATE 客户端本地使用的ip地址和端口号，代理服务器可以用这个信息对访问进行一些限制
	var target string
	if buf[3] == uint8(1) { // IP,如果客户端请求的是域名，到这里也是IP
		var targetBuf [6]byte
		_, err = reader.Read(targetBuf[:])
		if err != nil {
			logger.Zap.Sugar().Info(err)
			return nil, err
		}
		//target = fmt.Sprintf("%s:%d", net.IPv4(targetBuf[0], targetBuf[1], targetBuf[2], targetBuf[3]).String(), binary.BigEndian.Uint16(targetBuf[4:6]))
		target = fmt.Sprintf("%v.%v.%v.%v:%d", targetBuf[0], targetBuf[1], targetBuf[2], targetBuf[3], binary.BigEndian.Uint16(targetBuf[4:6]))
	} else if buf[3] == uint8(3) { // 域名,域名类型，DST.ADDR的第一个字节是长度
		var domainLengthBuf [1]byte
		_, err = reader.Read(domainLengthBuf[:])
		if err != nil {
			logger.Zap.Sugar().Info(err)
			return nil, err
		}
		var targetBufLength = int(domainLengthBuf[0]) + 2
		var targetBuf = make([]byte, targetBufLength)
		_, err = reader.Read(targetBuf[:targetBufLength])
		if err != nil {
			logger.Zap.Sugar().Info(err)
			return nil, err
		}
		target = fmt.Sprintf("%s:%d", string(targetBuf[:targetBufLength-2]), binary.BigEndian.Uint16(targetBuf[targetBufLength-2:targetBufLength]))
	} else { // IPV6等其他暂不支持
		//response = ResponseCodeSeven
	}
	log.Printf("客户端:%s，请求：%s\n", conn.RemoteAddr().String(), target)

	cmd := buf[1]
	// CONNECT 连接目标服务器
	if cmd == uint8(1) {
		resp := uint8(0)
		msg := ""
		connT, err := net.Dial("tcp", target)
		if err != nil {
			logger.Zap.Sugar().Error(err, " ", target)
			msg = err.Error()
			resp = ErrHostUnreachable
			if strings.Contains(msg, "refused") {
				resp = ErrConnectionRefused
			} else if strings.Contains(msg, "network is unreachable") {
				resp = ErrNetworkUnreachable
			}
		}

		res := []byte{Socks5Version, resp, uint8(0), uint8(1)}
		if resp != uint8(0) {
			res = append(res, []byte{0, 0, 0, 0}...) // IP
			res = append(res, []byte{0, 0}...)       // Port
			_, err = conn.Write(res)
			return nil, errors.New(msg)
		}
		if client, ok := connT.LocalAddr().(*net.TCPAddr); ok {
			port := binary.BigEndian.AppendUint16([]byte(nil), uint16(client.Port))
			res = append(res, client.IP...)
			res = append(res, port...)
		}

		_, err = conn.Write(res)

		return connT, err
	}

	return nil, nil
}

func readMethods(reader *bufio.Reader) ([]byte, error) {
	methodSize, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("read methodSize failed:%w", err)
	}

	methods := make([]byte, methodSize)
	readerL := io.LimitReader(reader, int64(methodSize))
	_, err = readerL.Read(methods)

	return methods, err
}
