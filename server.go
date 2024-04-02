package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"socks5/logger"
	"socks5/protocol"
)

type Server struct {
	addr string
	port int
}
type Config struct{}

var (
	Socks5Version = uint8(5)
	VersionError  = errors.New("version error")
	MethodError   = errors.New("method error")
)

func NewServer(addr string, opts ...Option) *Server {
	var options Options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			panic(err)
		}
	}

	var port int
	if options.port == nil {
		port = DefaultPort
	} else {
		if *options.port == 0 {
			port = DefaultPort
		} else {
			port = *options.port
		}
	}

	return &Server{
		addr: addr,
		port: port,
	}
}

func (s *Server) ListenAndServe() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.addr, s.port))
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
	log.Printf("%s connected \n", conn.RemoteAddr())
	bufConn := bufio.NewReader(conn) // 可以多次读，io.reader读一次就没了
	var socks5 *protocol.Socks5
	if socks5 == nil {
		socks5 = protocol.NewSocks5()
	}

	/// 认证过程
	// 1 客户端请求认证（检查版本，选定认证方式）, 2 服务器返回选定的认证方法
	err := socks5.CheckVersionAndAuthMethod(bufConn, conn)
	if err != nil {
		logger.Zap.Sugar().Info(err)
		return
	}
	// 3 如果为账号密码认证客户端再次发送账密密码进行认证 ,4 服务器响应账号密码认证结果. 如果无需账号密码认证，则直接跳过此步骤
	if socks5.Method == protocol.AuthMethodUsernamePwd {
		err = socks5.Auth(bufConn, conn)
		if err != nil {
			logger.Zap.Sugar().Info(err)
			return
		}
	}

	/// 命令过程
	Tconn, err := socks5.CreateProxy(conn, conn)
	if err != nil {
		return
	}

	/// 数据转发
	protocol.Forward(conn, Tconn)
}
