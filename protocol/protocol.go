package protocol

import (
	"errors"
	"io"
	"net"
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

func (s *Socks5) CheckVersion(reader io.Reader) error {
	version := []byte{0}
	_, err := reader.Read(version)
	if err != nil {
		return err
	}
	if version[0] != Socks5Version {
		return VersionError
	}

	return nil
}

func (s *Socks5) CheckMethod(reader io.Reader, writer io.Writer) error {
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
			break
		}
	}

	s.Method = method
	_, err = writer.Write([]byte{Socks5Version, method})
	return err
}

func (s *Socks5) Check(reader io.Reader, writer io.Writer) error {
	//TODO

	return nil
}
func (s *Socks5) CreateProxy(reader io.Reader, writer io.Writer) error {
	//var response = ResponseCodeZero

	//var buf [256]byte
	//n, err := reader.Read(buf[:])
	//cmd := buf[1]
	//port := buf[n-2 : n]
	//
	//
	//var target string
	//if buf[3] == 0x01 { // IP
	//	addr := buf[4 : n-2]
	//	target = fmt.Sprintf("%s:%d", net.IPv4(addr[0], addr[1], addr[2], addr[3]).String(), binary.BigEndian.Uint16(port))
	//} else if buf[3] == 0x03 { // 域名,域名类型，DST.ADDR的第一个字节是长度
	//	target = fmt.Sprintf("%s:%d", string(buf[5:n-2]), binary.BigEndian.Uint16(port))
	//} else { // IPV6等其他暂不支持
	//	response = 0x07
	//}
	//
	//ipSli, portSli, pConn, err := process(target)
	//if err != nil {
	//	log.Logger.Sugar().Info(err)
	//	continue
	//}
	//socks5.Conn = pConn
	//socks5.BndPort = portSli
	//socks5.BndAddr = ipSli
	//
	//// 发回客户端
	//sendData := make([]byte, len(ipSli)+6)
	//sendData[0] = 0x05
	//sendData[1] = byte(response)
	//sendData[2] = 0x00
	//sendData[3] = 0x01
	//sendData = append(sendData, ipSli...)
	//sendData = append(sendData, portSli...)
	//if _, err = conn.Write(sendData); err != nil {
	//	log.Logger.Sugar().Info(err)
	//	return
	//}

	return nil
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

func (s *Socks5) Forward(source net.Conn) {
	forward := func(src, dest net.Conn) {
		defer src.Close()
		defer dest.Close()
		io.Copy(src, dest)
	}
	go forward(source, s.Conn)
	go forward(s.Conn, source)
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
