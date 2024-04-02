package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"socks5/logger"
	"strconv"
	"strings"
	"sync"
)

var (
	Socks5Version = uint8(5)
	VersionError  = errors.New("version error")
	MethodError   = errors.New("method error")

	ResponseCodeZero  = uint8(0) // 代理服务器连接目标服务器成功
	ResponseCodeOne   = uint8(1) // 代理服务器故障
	ResponseCodeTwo   = uint8(2) // 代理服务器规则集不允许连接
	ResponseCodeThree = uint8(3) // 网络无法访问
	ResponseCodeFour  = uint8(4) // 目标服务器无法访问（主机名无效）
	ResponseCodeFive  = uint8(5) // 连接目标服务器被拒绝
	ResponseCodeSix   = uint8(6) // TTL已过期
	ResponseCodeSeven = uint8(7) // 不支持的命令
	ResponseCodeEight = uint8(8) // 不支持的目标服务器地址类型
)

//type ProxyData struct {
//	DataType int
//	Data     []byte
//	Error    error
//}

type Socks5 struct {
	Mu          sync.RWMutex
	Ver         int64    // 版本
	CurrentStep int      // 步骤
	Method      uint8    // 验证方式
	DstAddr     []byte   // 目标地址
	DstPort     []byte   // 目标端口号
	BndAddr     []byte   // 代理服务器连接目标服务器成功后的代理服务器 IP
	BndPort     []byte   // 代理服务器连接目标服务器成功后的代理服务器端口
	atyp        int      // 目标地址类型 0x01 IPv4，0x03 域名，0x04 IPv6
	Conn        net.Conn // 目标地址conn
}

func NewSocks5() *Socks5 {
	return &Socks5{Mu: sync.RWMutex{}}
}

func (s *Socks5) CheckVersionAndAuthMethod(reader io.Reader, writer io.Writer) error {
	// 检查版本
	version := []byte{0}
	_, err := reader.Read(version)
	if err != nil {
		return err
	}
	if version[0] != Socks5Version {
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
			s.Method = m
			break
		} else if m == AuthMethodUsernamePwd {
			s.Method = m
			err = s.Check(reader, writer)
			if err != nil {
				logger.Zap.Sugar().Info(err)
				return err
			}
			break
		}
	}

	s.Method = method
	_, err = writer.Write([]byte{Socks5Version, method})
	return err
}

func (s *Socks5) Auth(reader io.Reader, writer io.Writer) error {
	status := uint8(0)

	err := s.Check(reader, writer)
	if err != nil {
		status = uint8(1) // 大于0认证失败
	}

	_, err = writer.Write([]byte{Socks5Version, status})

	return err
}

func (s *Socks5) Check(reader io.Reader, writer io.Writer) error {
	//TODO

	return nil
}
func (s *Socks5) CreateProxy(reader io.Reader, writer io.Writer) (net.Conn, error) {
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
	log.Println(target)

	cmd := buf[1]
	// CONNECT 连接目标服务器
	if cmd == uint8(1) {
		conn, err := net.Dial("tcp", target)
		if err != nil {
			logger.Zap.Sugar().Info(err)
			return nil, err
		}
		ipSlic := strings.Split(conn.LocalAddr().String(), ":")
		ip := net.ParseIP(ipSlic[0])
		var port []byte

		num, err := strconv.ParseUint(ipSlic[1], 10, 16)
		if err != nil {
			logger.Zap.Sugar().Info(err)
			return nil, err
		}
		uint16Num := uint16(num)
		port = binary.BigEndian.AppendUint16(port, uint16Num)
		res := []byte{Socks5Version, uint8(0), uint8(0)}
		res = append(res, ip.To4()...)
		res = append(res, port...)
		_, err = writer.Write(res)

		return conn, err
	}

	return nil, nil
}

func readMethods(r io.Reader) ([]byte, error) {
	header := []byte{0}
	if _, err := r.Read(header); err != nil {
		return nil, err
	}

	numMethods := int(header[0])
	methods := make([]byte, numMethods)
	_, err := io.ReadAtLeast(r, methods, numMethods)
	return methods, err
}

func Forward(source, dest net.Conn) {
	forward := func(src, dest net.Conn) {
		defer src.Close()
		defer dest.Close()
		io.Copy(src, dest)
	}
	go forward(source, dest)
	go forward(dest, source)
}

// GetUsernameAnePwd 连接的第三步，验证用户名密码
// VERSION		USERNAME_LENGTH		USERNAME	PASSWORD_LENGTH		PASSWORD
// 1字节			1字节				1-255字节	1字节				1-255字节
// 0x01			0x01				……			0x01				……
//func (s *Socks5) GetUsernameAnePwd(clientData []byte) (string, string, error) {
//	var index int32 = 1
//	usernameLen, err := binary.ReadInt32FromBinary(clientData[index : index+1])
//	if err != nil {
//		return "", "", err
//	}
//
//	index += 1
//	username := string(clientData[index : index+usernameLen])
//	index += usernameLen
//
//	pwdLen, err := binary.ReadInt32FromBinary(clientData[index : index+1])
//	if err != nil {
//		return "", "", err
//	}
//	index += 1
//	pwd := string(clientData[index : index+pwdLen])
//
//	return username, pwd, nil
//}

// GetCmd 连接的第三步，验证用户名密码
// VERSION		COMMAND		RSV		ADDRESS_TYPE	DST.ADDR	DST.PORT
// 1字节			1字节		1字节	1字节			可变成长度	2字节
//func (s *Socks5) GetCmd(clientData []byte) (int32, int32, int64, int64, error) {
//	var index int32 = 1
//	cmd, err := binary.ReadInt32FromBinary(clientData[index : index+1])
//	if err != nil {
//		return 0, 0, 0, 0, err
//	}
//	index += 2
//	addressType, err := binary.ReadInt32FromBinary(clientData[index : index+1])
//	if err != nil {
//		return 0, 0, 0, 0, err
//	}
//	index += 1
//
//	addr, n := bin.Varint(clientData[index:])
//
//	index += int32(n)
//	port, _ := bin.Varint(clientData[index : index+2])
//
//	return int32(clientData[1]), int32(clientData[3]), addr, int64(clientData[len(clientData)-2,len(clientData)]), nil
//}
