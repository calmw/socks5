package main

import (
	"io"
	"log"
	"net/http"

	"golang.org/x/net/proxy"
)

func main() {

	var auth proxy.Auth
	auth.User = "cisco"
	auth.Password = "123456"
	dialer, _ := proxy.SOCKS5("tcp", "127.0.0.1:6666", &auth, proxy.Direct)

	client := &http.Client{
		Transport: &http.Transport{Dial: dialer.Dial},
	}

	//resp, err := client.Get("https://ip.gs")
	resp, err := client.Get("https://www.baidu.com/")
	if err != nil {
		log.Println(err, 111)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	println(string(bodyBytes))

}
