package main

import (
	"github.com/calmw/socks5"
)

func main() {
	server := socks5.NewServer("0.0.0.0", socks5.WithPort(6666))
	server.ListenAndServe()
}
