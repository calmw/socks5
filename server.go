package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"socks5/protocol"
)

var connSet map[*net.Conn]*protocol.Socks5

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:6666")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	var conn net.Conn
	for {
		conn, err = listener.Accept()
		if err == io.EOF { // 处理断开的链接
			_, ok := connSet[&conn]
			if ok {
				delete(connSet, &conn)
				continue
			}
		} else if err != nil {
			log.Println(err)
			continue
		}

		var content []byte
		_, err = conn.Read(content)
		socks5 := connSet[&conn]
		if connSet[&conn] == nil { // 第一次连接
			socks5 = protocol.NewSocks5()

			// 第一步:处理客户端发送的报头
			err = socks5.CheckValidateType(content)
			if err != nil {
				log.Println(err)
				continue
			}

			// 第二步:代理服务器响应报头
			sendData := make([]byte, 2)
			binary.PutVarint(sendData[:1], socks5.Ver)
			binary.PutVarint(sendData[1:2], socks5.Method)
			if _, err = conn.Write(sendData); err != nil {
				log.Println(err)
				continue
			}
			socks5.Step = 2

			connSet[&conn] = socks5
		} else {
			if socks5.Step == 2 {
				var status int64 = 0
				if socks5.Method == 0x02 {
					username, pwd, err := socks5.GetUsernameAnePwd(content)
					if err != nil {
						log.Println(err)
						continue
					}

					// 检查用户名密码
					fmt.Println(username, pwd)
					//TODO
					//检查失败返回status>0
				}
				sendData := make([]byte, 2)
				binary.PutVarint(sendData[:1], socks5.Ver)
				binary.PutVarint(sendData[1:2], status)
				if _, err = conn.Write(sendData); err != nil {
					log.Println(err)
					continue
				}
				socks5.Step = 3
			} else if socks5.Step == 3 {
				cmd, addressType, addr, port, err := socks5.GetCmd(content)
				if err != nil {
					log.Println(err)
					continue
				}

				// 发回客户端
				//sendData := make([]byte, 2)
				//binary.PutVarint(sendData[:1], socks5.Ver)
				//binary.PutVarint(sendData[1:2], status)
				//if _, err = conn.Write(sendData); err != nil {
				//	log.Println(err)
				//	continue
				//}
				// 代理流量
				fmt.Println(cmd, addressType, addr, port)

			}
		}
	}

}
