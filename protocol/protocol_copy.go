package protocol

//
//import (
//	"errors"
//	"fmt"
//	"io"
//	"net"
//	"sync"
//)
//
//var (
//	Socks5Version = uint8(5)
//	VersionError  = errors.New("version error")
//	MethodError   = errors.New("method error")
//)
//
////type ProxyData struct {
////	DataType int
////	Data     []byte
////	Error    error
////}
//
//type Socks5 struct {
//	Mu          sync.RWMutex
//	Ver         int64    // 版本
//	CurrentStep int      // 步骤
//	Method      int64    // 加密方式
//	DstAddr     []byte   // 目标地址
//	DstPort     []byte   // 目标端口号
//	BndAddr     []byte   // 代理服务器连接目标服务器成功后的代理服务器 IP
//	BndPort     []byte   // 代理服务器连接目标服务器成功后的代理服务器端口
//	atyp        int      // 目标地址类型 0x01 IPv4，0x03 域名，0x04 IPv6
//	Conn        net.Conn // 目标地址conn
//}
//
//func NewSocks5() *Socks5 {
//	return &Socks5{Mu: sync.RWMutex{}}
//}
//
//// CheckVersion 连接的第一步
//func (s *Socks5) CheckVersion(reader io.Reader) error {
//	// 版本检测
//	version := []byte{0}
//	_, err := io.LimitReader(reader, 0).Read(version)
//	if err != nil {
//		return err
//	}
//	if version[0] != Socks5Version {
//		return VersionError
//	}
//
//	return nil
//}
//
//func (s *Socks5) CheckMethod(reader io.Reader, writer io.Writer) error {
//	fmt.Println(1)
//	methods, err := readMethods(reader)
//	if err != nil {
//		return err
//	}
//	fmt.Println(2)
//
//	var method uint8 = AuthMethodUnSupport
//	for _, m := range methods {
//		if m == AuthMethodNo {
//			method = m
//			break
//		} else if m == AuthMethodUsernamePwd {
//			method = m
//			break
//		}
//	}
//	fmt.Println(6)
//
//	_, err = writer.Write([]byte{Socks5Version, method})
//	return err
//}
//
//func readMethods(r io.Reader) ([]byte, error) {
//	header := []byte{0}
//	fmt.Println(4)
//	if _, err := r.Read(header); err != nil {
//		return nil, err
//	}
//
//	fmt.Println(3)
//	numMethods := int(header[0])
//	methods := make([]byte, numMethods)
//	_, err := io.ReadAtLeast(r, methods, numMethods)
//	return methods, err
//}
//
//func (s *Socks5) CheckValidateType(clientData []byte) error {
//	// 版本检测
//	if clientData[0] != 0x05 {
//		return VersionError
//	}
//	s.Ver = 0x05
//
//	// 方法检测
//	methods := clientData[2:]
//	for _, method := range methods {
//		if method == 0x00 {
//			s.Method = 0x00
//			return nil
//		} else if method == 0x02 {
//			s.Method = 0x02
//			return nil
//		}
//	}
//	if s.Method != 0x00 && s.Method != 0x02 {
//		return MethodError
//	}
//
//	s.CurrentStep = 1
//	return nil
//}
//
//func (s *Socks5) Forward(source net.Conn) {
//	forward := func(src, dest net.Conn) {
//		defer src.Close()
//		defer dest.Close()
//		io.Copy(src, dest)
//	}
//	go forward(source, s.Conn)
//	go forward(s.Conn, source)
//}
//
//// GetUsernameAnePwd 连接的第三步，验证用户名密码
//// VERSION		USERNAME_LENGTH		USERNAME	PASSWORD_LENGTH		PASSWORD
//// 1字节			1字节				1-255字节	1字节				1-255字节
//// 0x01			0x01				……			0x01				……
////func (s *Socks5) GetUsernameAnePwd(clientData []byte) (string, string, error) {
////	var index int32 = 1
////	usernameLen, err := binary.ReadInt32FromBinary(clientData[index : index+1])
////	if err != nil {
////		return "", "", err
////	}
////
////	index += 1
////	username := string(clientData[index : index+usernameLen])
////	index += usernameLen
////
////	pwdLen, err := binary.ReadInt32FromBinary(clientData[index : index+1])
////	if err != nil {
////		return "", "", err
////	}
////	index += 1
////	pwd := string(clientData[index : index+pwdLen])
////
////	return username, pwd, nil
////}
//
//// GetCmd 连接的第三步，验证用户名密码
//// VERSION		COMMAND		RSV		ADDRESS_TYPE	DST.ADDR	DST.PORT
//// 1字节			1字节		1字节	1字节			可变成长度	2字节
////func (s *Socks5) GetCmd(clientData []byte) (int32, int32, int64, int64, error) {
////	var index int32 = 1
////	cmd, err := binary.ReadInt32FromBinary(clientData[index : index+1])
////	if err != nil {
////		return 0, 0, 0, 0, err
////	}
////	index += 2
////	addressType, err := binary.ReadInt32FromBinary(clientData[index : index+1])
////	if err != nil {
////		return 0, 0, 0, 0, err
////	}
////	index += 1
////
////	addr, n := bin.Varint(clientData[index:])
////
////	index += int32(n)
////	port, _ := bin.Varint(clientData[index : index+2])
////
////	return int32(clientData[1]), int32(clientData[3]), addr, int64(clientData[len(clientData)-2,len(clientData)]), nil
////}
