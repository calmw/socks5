package main

import (
	"bufio"
	"errors"
	"net"
	"socks5/pkg/log"
	"socks5/protocol"
)

type Server struct{}
type Config struct{}

var (
	Socks5Version = uint8(5)
	VersionError  = errors.New("version error")
	MethodError   = errors.New("method error")
)

func New() (*Server, error) {
	return &Server{}, nil
}

func (s *Server) ListenAndServe(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	var conn net.Conn
	for {
		conn, err = listener.Accept()
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	bufConn := bufio.NewReader(conn) // 可以多次读，io.reader读一次就没了
	var socks5 *protocol.Socks5
	if socks5 == nil {
		socks5 = protocol.NewSocks5()
	}
	// 检查版本
	err := socks5.CheckVersion(bufConn)
	if err != nil {
		log.Logger.Sugar().Info(err)
		return
	}
	// 检查验证方式
	err = socks5.CheckMethod(bufConn, conn)
	if err != nil {
		log.Logger.Sugar().Info(err)
		return
	}
	// 如果不是用户名-密码验证方式就直接下一步
	if socks5.Method == protocol.AuthMethodUsernamePwd {
		err = socks5.Check(bufConn, conn)
		if err != nil {
			log.Logger.Sugar().Info(err)
			return
		}
	}
	//

}
