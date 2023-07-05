package main

import (
	"RPC"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	//addr := "0.0.0.0:9999"

	var centerAddr string
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Failed to get hostname:", err)
		return
	}

	// 获取主机的IP地址
	addrs, err := net.LookupIP(hostname)
	if err != nil {
		fmt.Println("Failed to get IP addresses for", hostname, ":", err)
		return
	}
	for _, addr := range addrs {
		centerAddr = addr.To4().String()
	}

	var addr string
	IP := flag.String("l", "0.0.0.0", "注册中心监听的IP地址")
	port := flag.String("p", "", "注册中心监听的端口号(必须输入)")

	flag.Parse()

	if *port == "" {
		flag.PrintDefaults()
		os.Exit(1)
	} else {
		addr = *IP + ":" + *port
	}

	flag.Parse()

	center := RPC.NewCenter(centerAddr+":"+*port, 10*time.Second)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	log.Println("center is on ready")

	for {
		conn, err := ln.Accept()
		log.Println(conn.RemoteAddr())
		if err != nil {
			log.Println("Error: ", err)
			continue
		}
		go center.Serve(conn)
	}
}
