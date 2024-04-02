package protocol

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"socks5/logger"
)

var (
	Socks5Version = uint8(5)
	VersionError  = errors.New("version error")
)

type Socks5 struct {
	AuthMethod uint8 // 验证方式
}

func NewSocks5() *Socks5 {
	return &Socks5{}
}

func (s *Socks5) CheckVersionAndAuthMethod(reader *bufio.Reader, conn net.Conn) error {
	// 检查版本
	version, err := reader.ReadByte()
	if err != nil {
		return err
	}
	if version != Socks5Version {
		return VersionError
	}

	// 认证方法
	methods, err := readMethods(reader)
	if err != nil {
		return err
	}

	var method uint8 = AuthMethodUnSupport
	for _, m := range methods {
		if m == AuthMethodNo {
			method = m
			break
		} else if m == AuthMethodUsernamePwd {
			method = m
			err = s.Check(reader, conn)
			if err != nil {
				logger.Zap.Sugar().Info(err)
				return err
			}
			break
		}
	}

	_, err = conn.Write([]byte{Socks5Version, method})
	return err
}

func (s *Socks5) Auth(reader io.Reader, conn net.Conn) error {
	status := uint8(0)

	err := s.Check(reader, conn)
	if err != nil {
		status = uint8(1) // 大于0认证失败
	}

	_, err = conn.Write([]byte{Socks5Version, status})

	return err
}

func (s *Socks5) Check(reader io.Reader, writer io.Writer) error {
	//TODO

	return nil
}

func (s *Socks5) CreateProxy(reader io.Reader, conn net.Conn) (net.Conn, error) {
	//var response = ResponseCodeZero

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
		target = fmt.Sprintf("%s:%d", net.IPv4(targetBuf[0], targetBuf[1], targetBuf[2], targetBuf[3]).String(), binary.BigEndian.Uint16(targetBuf[4:6]))
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
		connT, err := net.Dial("tcp", target)
		if err != nil {
			logger.Zap.Sugar().Info(err)
			return nil, err
		}

		res := []byte{Socks5Version, uint8(0), uint8(0), uint8(1)}
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
	_, err = io.ReadFull(reader, methods)
	if err != nil {
		return nil, fmt.Errorf("read method failed:%w", err)
	}

	return methods, err
}
