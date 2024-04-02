#### socks5 介绍

- 介绍
    - SOCKS5 是一个代理协议，它在使用TCP/IP协议通讯的前端机器和服务器机器之间扮演一个中介角色，使得内部网中的前端机器变得能够访问Internet网中的服务器，或者使通讯更加安全。SOCKS5
      服务器通过将前端发来的请求转发给真正的目标服务器，
      模拟了一个前端的行为。在这里，前端和SOCKS5之间也是通过TCP/IP协议进行通讯，前端将原本要发送给真正服务器的请求发送给SOCKS5服务器，然后SOCKS5服务器将请求转发给真正的服务器。
- 文档：https://www.ietf.org/rfc/rfc1928.txt
- 参考：[socks5 协议详解.pdf](..%2Fstatic%2Fsocks5%20%E5%8D%8F%E8%AE%AE%E8%AF%A6%E8%A7%A3.pdf)
- 优缺点：[SOCKS5 代理有哪些优点.pdf](..%2Fstatic%2FSOCKS5%20%E4%BB%A3%E7%90%86%E6%9C%89%E5%93%AA%E4%BA%9B%E4%BC%98%E7%82%B9.pdf)