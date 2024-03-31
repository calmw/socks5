package main

//
//import (
//	"bufio"
//	"encoding/binary"
//	"fmt"
//	"io"
//	"net"
//	"socks5/pkg/log"
//	"socks5/protocol"
//	"strconv"
//	"strings"
//)
//
//func main() {
//	listener, err := net.Listen("tcp", "0.0.0.0:6666")
//	if err != nil {
//		panic(err)
//	}
//	defer listener.Close()
//
//	var conn net.Conn
//	for {
//		conn, err = listener.Accept()
//
//		go handleConn(conn)
//	}
//}
//
//func handleConn(conn net.Conn) {
//	defer func() {
//		fmt.Printf("client [%s] close\n", conn.RemoteAddr().String())
//		conn.Close()
//	}()
//
//	var buf [1024]byte
//	var socks5 *protocol.Socks5
//	for {
//		n, err := conn.Read(buf[:])
//		if err == io.EOF { // 处理断开的链接
//			return
//		} else if err != nil {
//			log.Logger.Sugar().Info(err)
//			return
//		}
//
//		//fmt.Println("read data :", n, buf[:n])
//
//		if socks5 == nil { // 第一次连接
//			socks5 = protocol.NewSocks5()
//
//			if n < 3 {
//				log.Logger.Sugar().Info("content length error:", n)
//				return
//			}
//
//			// 第一步:处理客户端发送的报头
//			err = socks5.CheckValidateType(buf[:n])
//			if err != nil {
//				log.Logger.Sugar().Info(err)
//				return
//			}
//			socks5.CurrentStep = 1
//
//			// 第二步:代理服务器响应报头
//			//sendData := make([]byte, 2)
//			//sendData[0] = byte(socks5.Ver)
//			//sendData[1] = byte(socks5.Method)
//			//if _, err = conn.Write(sendData); err != nil {
//			//	log.Logger.Sugar().Info(err)
//			//	return
//			//}
//
//			bufConn := bufio.NewReader(conn)
//			err = protocol.CheckValidateMethod(bufConn, conn)
//			if err != nil {
//				log.Logger.Sugar().Info(err)
//				return
//			}
//
//			// 如果选择0x00不加密方式，客户端将跳过发送用户名密码步骤
//			if socks5.Method == 0x00 {
//				socks5.CurrentStep = 3
//			} else {
//				socks5.CurrentStep = 2
//			}
//		} else {
//			if socks5.CurrentStep == 2 {
//				status := 0x00
//				// 验证用户名密码，并将结果写回客户端
//				if socks5.Method == 0x02 {
//					// 检查用户名密码,检查失败返回status>0
//					//TODO
//				}
//
//				sendData := make([]byte, 2)
//				sendData[0] = 0x05
//				sendData[1] = byte(status)
//				if _, err = conn.Write(sendData); err != nil {
//					log.Logger.Sugar().Info(err)
//					return
//				}
//				socks5.CurrentStep = 3
//			} else if socks5.CurrentStep == 3 {
//				var response = 0x00
//
//				cmd := buf[1]
//				port := buf[n-2 : n]
//
//				if cmd != 0x01 {
//					response = 0x07
//				}
//				var target string
//				if buf[3] == 0x01 { // IP
//					addr := buf[4 : n-2]
//					target = fmt.Sprintf("%s:%d", net.IPv4(addr[0], addr[1], addr[2], addr[3]).String(), binary.BigEndian.Uint16(port))
//				} else if buf[3] == 0x03 { // 域名,域名类型，DST.ADDR的第一个字节是长度
//					target = fmt.Sprintf("%s:%d", string(buf[5:n-2]), binary.BigEndian.Uint16(port))
//				} else { // IPV6等其他暂不支持
//					response = 0x07
//				}
//
//				ipSli, portSli, pConn, err := process(target)
//				if err != nil {
//					log.Logger.Sugar().Info(err)
//					continue
//				}
//				socks5.Conn = pConn
//				socks5.BndPort = portSli
//				socks5.BndAddr = ipSli
//
//				// 发回客户端
//				sendData := make([]byte, len(ipSli)+6)
//				sendData[0] = 0x05
//				sendData[1] = byte(response)
//				sendData[2] = 0x00
//				sendData[3] = 0x01
//				sendData = append(sendData, ipSli...)
//				sendData = append(sendData, portSli...)
//				if _, err = conn.Write(sendData); err != nil {
//					log.Logger.Sugar().Info(err)
//					return
//				}
//				socks5.CurrentStep = 4
//			} else if socks5.CurrentStep == 4 {
//				fmt.Println(998765)
//				go func() {
//					_, err := io.Copy(socks5.Conn, conn)
//					fmt.Println(err, 1)
//				}()
//
//				aaa, err := io.Copy(conn, socks5.Conn)
//
//				fmt.Println(aaa, err, 2)
//				//socks5.Forward(conn)
//
//				//_, err = socks5.Conn.Write(buf[:n])
//				//if err != nil {
//				//	log.Logger.Sugar().Info(err)
//				//	return
//				//}
//				//var bufR [1024]byte
//
//				//fmt.Println(conn.LocalAddr(), conn.RemoteAddr(), socks5.Conn.RemoteAddr())
//				//fmt.Println(conn.Write([]byte("ssss")))
//				//sg := &sync.WaitGroup{}
//				//sg.Add(1)
//				//go func(c net.Conn) {
//				//	for {
//				//		nt, err := socks5.Conn.Read(bufR[:])
//				//		if err == io.EOF { // 处理断开的链接
//				//			log.Logger.Sugar().Info(err)
//				//			break
//				//		} else if err != nil {
//				//			log.Logger.Sugar().Info(err)
//				//			break
//				//		}
//				//		//targetData = append(targetData, bufR[:nt]...)
//				//		num, err := c.Write(bufR[:nt])
//				//		fmt.Println(num, err)
//				//		if err != nil {
//				//			log.Logger.Sugar().Info(err)
//				//		}
//				//
//				//		if nt < 1024 {
//				//			break
//				//		}
//				//
//				//	}
//				//
//				//	sg.Done()
//				//}(conn)
//				//sg.Wait()
//			}
//		}
//	}
//
//}
//
//func process(target string) ([]byte, []byte, net.Conn, error) {
//	conn, err := net.Dial("tcp", target)
//	if err != nil {
//		return nil, nil, nil, err
//	}
//	ipSlic := strings.Split(conn.LocalAddr().String(), ":")
//	ip := net.ParseIP(ipSlic[0])
//	var port []byte
//
//	num, err := strconv.ParseUint(ipSlic[1], 10, 16)
//	if err != nil {
//		return nil, nil, nil, err
//	}
//	uint16Num := uint16(num)
//	port = binary.BigEndian.AppendUint16(port, uint16Num)
//
//	return ip.To4(), port, conn, err
//}
