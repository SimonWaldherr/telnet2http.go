package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func handleConnection(c net.Conn, msgchan chan<- string) {
	buf := make([]byte, 4096)
	for {
		n, err := c.Read(buf)
		if (err != nil) || (n == 0) {
			c.Close()
			break
		}
		msgchan <- string(buf[0:n])
	}
	fmt.Printf("Connection from %v closed.\n", c.RemoteAddr())
}

func printMessages(msgchan <-chan string, urlstr string, method string) {
	for {
		msg := strings.TrimSpace(<-msgchan)
		fmt.Printf("data: %s\n", msg)
		data := url.Values{}
		data.Add("value", strings.TrimSpace(msg))
		client := &http.Client{}
		if method == "POST" {
			r, _ := http.NewRequest("POST", urlstr, bytes.NewBufferString(data.Encode()))
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
			client.Do(r)
		} else {
			r, _ := http.NewRequest("GET", urlstr+"?"+data.Encode(), nil)
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
			client.Do(r)
		}
	}
}

func main() {
	flag.Parse()
	url := flag.Arg(0)
	port := ":" + flag.Arg(1)
	method := flag.Arg(2)
	if url == "" {
		url = "localhost"
	}
	if port == ":" {
		port = ":23"
	}
	if method == "" {
		method = "POST"
	}
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	msgchan := make(chan string)
	go printMessages(msgchan, url, method)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConnection(conn, msgchan)
	}
}
