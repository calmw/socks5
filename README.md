# socks5

- 目前实现了以下三个功能中的第一个，可以满足大部分代理需求，后续功能开发中
    - 0x01 CONNECT 连接目标服务器
    - 0x02 BIND 绑定，客户端会接收来自代理服务器的链接，也就是说告诉代理服务器创建socket，监听来自目标机器的连接。像FTP服务器这种主动连接客户端的应用场景。
    - 0x03 UDP ASSOCIATE UDP中继

## 测试

- 终端代理命令：``` export all_proxy=socks5://127.0.0.1:6666 ```
    - 终端测试：![终端测试.png](static%2F%E7%BB%88%E7%AB%AF%E6%B5%8B%E8%AF%95.png)
- 系统设置：
    - ![配置.png](static%2F%E9%85%8D%E7%BD%AE.png)
    - 浏览器测试：
        - ![浏览器测试.png](static%2F%E6%B5%8F%E8%A7%88%E5%99%A8%E6%B5%8B%E8%AF%95.png)

#### 使用
- 使用示例：
``` go
package main

import (
	"github.com/calmw/socks5"
)

func main() {
	server := socks5.NewServer("0.0.0.0", socks5.WithPort(6666))
	server.ListenAndServe()
}
```