package protocol

import (
	bin "encoding/binary"
	"errors"
	"socks5/pkg/binary"
)

var (
	VersionError = errors.New("version error")
	MethodError  = errors.New("Method error")
)

type Socks5 struct {
	Ver     int64  // 版本
	Step    int    // 步骤
	Method  int64  // 加密方式
	dstAddr string // 目标地址
	dstPort int    // 目标端口号
	atyp    int    // 目标地址类型 0x01 IPv4，0x03 域名，0x04 IPv6
}

func NewSocks5() *Socks5 {
	return &Socks5{}
}

// CheckValidateType 连接的第一步
func (s *Socks5) CheckValidateType(clientData []byte) error {
	// 版本检测
	if version, err := binary.ReadInt32FromBinary(clientData[:1]); err != nil {
		return err
	} else if version != 0x05 {
		return VersionError
	}
	s.Ver = 0x05

	// 方法检测
	methods := clientData[2 : len(clientData)-1]
	for _, methodByte := range methods {
		method, err := binary.ReadInt32FromBinary([]byte{methodByte})
		if err != nil {
			return err
		}
		if method == 0x00 {
			s.Method = 0x00
			return nil
		} else if method == 0x02 {
			s.Method = 0x02
			return nil
		}
	}
	if s.Method != 0x00 && s.Method != 0x02 {
		return MethodError
	}

	s.Step = 1
	return nil
}

// GetUsernameAnePwd 连接的第三步，验证用户名密码
// VERSION		USERNAME_LENGTH		USERNAME	PASSWORD_LENGTH		PASSWORD
// 1字节			1字节				1-255字节	1字节				1-255字节
// 0x01			0x01				……			0x01				……
func (s *Socks5) GetUsernameAnePwd(clientData []byte) (string, string, error) {
	var index int32 = 1
	usernameLen, err := binary.ReadInt32FromBinary(clientData[index : index+1])
	if err != nil {
		return "", "", err
	}

	index += 1
	username := string(clientData[index : index+usernameLen])
	index += usernameLen

	pwdLen, err := binary.ReadInt32FromBinary(clientData[index : index+1])
	if err != nil {
		return "", "", err
	}
	index += 1
	pwd := string(clientData[index : index+pwdLen])

	return username, pwd, nil
}

// GetCmd 连接的第三步，验证用户名密码
// VERSION		COMMAND		RSV		ADDRESS_TYPE	DST.ADDR	DST.PORT
// 1字节			1字节		1字节	1字节			可变成长度	2字节
func (s *Socks5) GetCmd(clientData []byte) (int32, int32, int64, int64, error) {
	var index int32 = 1
	cmd, err := binary.ReadInt32FromBinary(clientData[index : index+1])
	if err != nil {
		return 0, 0, 0, 0, err
	}
	index += 2
	addressType, err := binary.ReadInt32FromBinary(clientData[index : index+1])
	if err != nil {
		return 0, 0, 0, 0, err
	}
	index += 1

	addr, n := bin.Varint(clientData[index:])

	index += int32(n)
	port, _ := bin.Varint(clientData[index : index+2])

	return cmd, addressType, addr, port, nil
}
