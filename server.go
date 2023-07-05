package RPC

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

type Server struct {
	Methods map[string]func(params ...interface{}) interface{}
	Timeout time.Duration
	Addr    string
}

func NewServer(timeout time.Duration, serverAddr string) *Server {
	server := Server{
		Methods: make(map[string]func(params ...interface{}) interface{}),
		Timeout: timeout,
		Addr:    serverAddr,
	}
	//server.CheckMethod()
	return &server
}

func (s *Server) GetAllMethods() []string {
	var keys []string
	for key := range s.Methods {
		keys = append(keys, key)
	}
	return keys
}

func (s *Server) Register(centerAddr string, method string, f func(params ...interface{}) interface{}) {
	s.Methods[method] = f

	var message Message

	conn, err := net.DialTimeout("tcp", centerAddr, s.Timeout)
	if err != nil {
		log.Println("dial failure " + err.Error())
		os.Exit(1)
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(conn.LocalAddr().String() + "conn close error" + err.Error())
		}
	}(conn)

	message.Type = 3
	message.Payload = map[string]interface{}{
		"methods": method,
		"addr":    s.Addr,
	}
	l, _ := json.Marshal(message.Payload)
	message.Length = uint16(len(l))
	Write(conn, message)
}

func (s *Server) Remove(method string) {
	delete(s.Methods, method)
}

func (s *Server) Serve(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("conn close error")
			return
		}
	}(conn)

	//设置读取超时时间
	if err := conn.SetReadDeadline(time.Now().Add(s.Timeout)); err != nil {
		return
	}

	var message Message
	var request Request
	message = Read(conn)
	log.Println(message)
	//log.Println("have read")
	if message.Type == 4 {
		r := message.Payload.(map[string]interface{})
		request.Method = r["method"].(string)
		if r["params"] == nil {
			request.Params = nil
		} else {
			request.Params = r["params"].([]interface{})
		}
	} else {
		return
	}

	f, ok := s.Methods[request.Method]
	if !ok {
		message.Type = 0
		message.Payload = errors.New(fmt.Sprintf("Method not found: %s", request.Method))
		p, _ := json.Marshal(message.Payload)
		message.Length = uint16(len(p))
		_ = Write(conn, message)
		return
	}

	//创建带有超时时间的上下文
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	//要设置chan的size
	ch := make(chan interface{}, 1)
	resultCh := make(chan interface{}, 1)

	//当f函数无法返回时，使用协程，避免阻塞，主线程超时会退出
	go func() {
		result := f(request.Params...)
		ch <- result
		resultCh <- result
	}()

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("calling method timeout")
			message.Type = 0
			message.Payload = "calling method timeout"
			message.Length = uint16(len(message.Payload.(string)))
			//log.Println(message.Length)
			_ = Write(conn, message)
		} else {
			log.Println(ctx.Err())
		}
	case <-ch:
		resp := Response{<-resultCh, ""}

		//设置写入超时时间
		if err := conn.SetWriteDeadline(time.Now().Add(s.Timeout)); err != nil {
			return
		}

		message.Type = 5
		message.Payload = resp
		respBytes, _ := json.Marshal(resp)
		message.Length = uint16(len(respBytes))
		message = Write(conn, message)
		if message.Type == 0 {
			log.Printf("server write request error")
			return
		}
	}
}

func (s *Server) KeepAlive(centerAddr string) {
	conn, err := net.DialTimeout("tcp", centerAddr, s.Timeout)
	if err != nil {
		log.Println("dial failure " + err.Error())
		os.Exit(1)
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(conn.LocalAddr().String() + "conn close error" + err.Error())
		}
	}(conn)

	var message Message
	message.Type = 2
	l, _ := json.Marshal(s.Addr)
	message.Length = uint16(len(l))
	message.Payload = s.Addr

	//设置写入超时时间
	if err := conn.SetWriteDeadline(time.Now().Add(s.Timeout)); err != nil {
		return
	}
	Write(conn, message)
}
