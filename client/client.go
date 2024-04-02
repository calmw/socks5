package main

import (
	"io"
	"net/http"

	"golang.org/x/net/proxy"
)

func main() {

	dialer, _ := proxy.SOCKS5("tcp", "127.0.0.1:6666", nil, proxy.Direct)

	client := &http.Client{

		Transport: &http.Transport{Dial: dialer.Dial},
	}

	//resp, _ := client.Get("http://ip.gs")
	resp, _ := client.Get("http://8.130.102.48:8000")

	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	println(string(bodyBytes))

}
