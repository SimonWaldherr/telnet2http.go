package main

//import needed libraries
import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	uri "net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

//EXAMPLE:
//go run telnet2http.go "destination url" [port [method [timeout]]]
//go run telnet2http.go "http://de.wikipedia.org:80/" 3000 GET 5
//telnet localhost 3000
//nc localhost 3000 < LICENSE

var timeout int64 = 0

//handle tcp/ip connections
func handleConnection(c *net.TCPConn, msgchan chan<- string) {
	defer c.Close()
	fmt.Printf("Connection from %v established.\n", c.RemoteAddr())
	if timeout != 0 {
		c.SetReadDeadline(time.Now().Add(time.Second * time.Duration(timeout)))
	}
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

//send message via http
func printMessages(msgchan <-chan string, urlstr string, method string) {
	var url string
	var body *bytes.Buffer
	var rtype string
	var clength string
	var bodyBytes []byte
	for {
		msg := strings.TrimSpace(<-msgchan)
		fmt.Printf("data: %s\n", msg)
		data := uri.Values{}
		data.Add("value", strings.TrimSpace(msg))
		client := &http.Client{}
		if method == "POST" {
			rtype = "POST"
			url = urlstr
			body = bytes.NewBufferString(data.Encode())
			clength = strconv.Itoa(len(data.Encode()))
		} else {
			rtype = "GET"
			url = urlstr + "?" + data.Encode()
			body = bytes.NewBufferString("")
			clength = "0"
		}
		fmt.Println(rtype + "\n" + url + "\n" + clength)
		r, _ := http.NewRequest(rtype, url, body)
		r.Header.Add("User-Agent", "telnet2http")
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Add("Content-Length", clength)
		rsp, err := client.Do(r)
		if err != nil {
			fmt.Println(err)
		} else {
			if rsp.StatusCode == 200 {
				bodyBytes, _ = ioutil.ReadAll(rsp.Body)
				fmt.Println(string(bodyBytes))
			} else if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("The remote end did not return a HTTP 200 (OK) response.")
			}
			rsp.Body.Close()
		}
	}
}

//main function to start listener
func main() {
	flag.Parse()
	url := flag.Arg(0)
	port := ":" + flag.Arg(1)
	method := flag.Arg(2)
	timeout, _ = strconv.ParseInt(flag.Arg(3), 10, 64)
	if url == "" {
		url = "localhost"
	}
	if port == ":" {
		port = ":23"
	}
	if method == "" {
		method = "POST"
	}
	fmt.Printf("Listening on port %v and sending via %v to %v\n", port, method, url)
	if timeout != 0 {
		fmt.Printf("Abort telnet connection after %v seconds\n", timeout)
	}
	l, err := net.Listen("tcp", port)
	ln := l.(*net.TCPListener)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	msgchan := make(chan string)
	go printMessages(msgchan, url, method)
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConnection(conn, msgchan)
	}
}
