package main

import (
	"RPC"
	"log"
	"net"
	"time"
)

func Add(params ...interface{}) interface{} {
	a := params[0].(float64)
	b := params[1].(float64)
	return a + b
}

func Loop(params ...interface{}) interface{} {
	select {}
}

func main() {
	server := RPC.NewServer(1 * time.Second)
	server.Register("Add", Add)
	server.Register("Loop", Loop)

	ln, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Println("Error:", err)
		return
	}
	log.Println("server is on ready")

	//var i = 1
	for {
		conn, err := ln.Accept()
		log.Println(conn.RemoteAddr())
		if err != nil {
			log.Println("Error:", err)
			continue
		}
		go server.Serve(conn)
		//log.Println(i)
		//i++
	}
}
