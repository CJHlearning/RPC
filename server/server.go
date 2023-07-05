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

func Add(params ...interface{}) interface{} {
	a := params[0].(float64)
	b := params[1].(float64)
	return a + b
}

func Sub(params ...interface{}) interface{} {
	a := params[0].(float64)
	b := params[1].(float64)
	return a - b
}

func Mul(params ...interface{}) interface{} {
	a := params[0].(float64)
	b := params[1].(float64)
	return a * b
}

func Div(params ...interface{}) interface{} {
	a := params[0].(float64)
	b := params[1].(float64)
	return a / b
}

func Square(params ...interface{}) interface{} {
	a := params[0].(float64)
	return a * a
}

func Big(params ...interface{}) interface{} {
	a := params[0].(float64)
	b := params[1].(float64)
	if a > b {
		return a
	} else {
		return b
	}
}

func Small(params ...interface{}) interface{} {
	a := params[0].(float64)
	b := params[1].(float64)
	if a < b {
		return a
	} else {
		return b
	}
}

func Equal(params ...interface{}) interface{} {
	a := params[0].(float64)
	b := params[1].(float64)
	return a == b
}

func HelloWorld(params ...interface{}) interface{} {
	return "Hello World!"
}

func Loop(params ...interface{}) interface{} {
	select {}
}

func main() {
	//serverAddr := "localhost:8888"
	//centerAddr := "localhost:9999"

	var serverAddr string
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
		serverAddr = addr.To4().String()
	}

	IP := flag.String("l", "0.0.0.0", "服务器监听的IP地址")
	port := flag.String("p", "", "服务器监听的端口号(必须输入)")
	centerAddr := flag.String("c", "", "注册中心地址(必须输入) 格式：'ip:port'")
	help := flag.Bool("h", false, "打印帮助参数")

	flag.Parse()

	var listenAddr string
	if *port == "" || *centerAddr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	} else {
		if *IP == "127.0.0.1" {
			*IP = "0.0.0.0"
		}
		listenAddr = *IP + ":" + *port
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	server := RPC.NewServer(1*time.Second, serverAddr+":"+*port)
	server.Register(*centerAddr, "Add", Add)
	server.Register(*centerAddr, "Sub", Sub)
	server.Register(*centerAddr, "Mul", Mul)
	server.Register(*centerAddr, "Div", Div)
	server.Register(*centerAddr, "Square", Square)
	server.Register(*centerAddr, "Big", Big)
	server.Register(*centerAddr, "Small", Small)
	server.Register(*centerAddr, "Equal", Equal)
	server.Register(*centerAddr, "HelloWorld", HelloWorld)
	server.Register(*centerAddr, "Loop", Loop)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	log.Println("server is on ready")

	go func(centerAddr string) {
		ticker := time.NewTicker(60 * time.Second) // 每60秒发送一次心跳包
		defer ticker.Stop()

		for range ticker.C {
			// 发送心跳包逻辑
			server.KeepAlive(centerAddr)
		}
	}(*centerAddr)

	for {
		conn, err := ln.Accept()
		log.Println(conn.RemoteAddr())
		if err != nil {
			log.Println("Error:", err)
			continue
		}
		go server.Serve(conn)
	}
}
