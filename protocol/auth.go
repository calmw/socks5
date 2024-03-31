package protocol

import (
	"errors"
)

var (
	UserAuthFailed  = errors.New("user authentication failed")
	NoSupportedAuth = errors.New("no supported authentication mechanism")

	// 这里只列举了部分认证方式，其他的参考文档
	AuthMethodNo          = uint8(0)   //不需要认证（常用）
	AuthMethodUsernamePwd = uint8(2)   // 账号密码认证（常用）
	AuthMethodUnSupport   = uint8(255) // 0xFF 无支持的认证方法
)

type AuthContext struct {
	Method  uint8             // 认证方法
	Payload map[string]string // 用于认证的数据，例如：用户名密码等数据
}
