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
	methods map[string]func(params ...interface{}) interface{}
	Timeout time.Duration
}

func NewServer(timeout time.Duration) *Server {
	server := Server{
		methods: make(map[string]func(params ...interface{}) interface{}),
		Timeout: timeout,
	}
	server.CheckMethod()
	return &server
}

func (s *Server) Register(method string, f func(params ...interface{}) interface{}) {
	s.methods[method] = f
}

func (s *Server) Remove(method string) {
	delete(s.methods, method)
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

	reqBytes := make([]byte, 1024)
	n, err := conn.Read(reqBytes)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			log.Println("conn read request timeout")
		} else {
			log.Println("conn read error: " + err.Error())
		}
		return
	}

	var req Request
	err = json.Unmarshal(reqBytes[:n], &req)
	if err != nil {
		log.Println("request Unmarshal error: " + err.Error())
		return
	}

	f, ok := s.methods[req.Method]
	if !ok {
		resp := Response{nil, fmt.Sprintf("Method not found: %s", req.Method)}
		respBytes, err := json.Marshal(resp)
		if err != nil {
			log.Println("response Marshal error:" + err.Error())
			return
		}

		_, err = conn.Write(respBytes)
		if err != nil {
			log.Println("conn write error: " + err.Error())
			return
		}
		return
	}

	//创建带有超时时间的上下文
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	//remember to set the value of channel's buffer size
	ch := make(chan interface{}, 1)
	resultCh := make(chan interface{}, 1)

	//use goroutine to avoid blocking when function f couldn't return a result
	go func() {
		result := f(req.Params...)
		ch <- result
		resultCh <- result
	}()

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("calling method timeout")
		} else {
			log.Println(ctx.Err())
		}
	case <-ch:
		resp := Response{<-resultCh, ""}
		respBytes, err := json.Marshal(resp)
		if err != nil {
			log.Println("response Marshal error: " + err.Error())
			return
		}

		//设置写入超时时间
		if err := conn.SetWriteDeadline(time.Now().Add(s.Timeout)); err != nil {
			return
		}

		_, err = conn.Write(respBytes)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				log.Println("conn write response timeout")
			} else {
				log.Println("conn write response error: " + err.Error())
			}
			return
		}
	}
}

func (s *Server) CheckMethod() {
	s.Register("CheckMethod", func(params ...interface{}) interface{} {
		var method string
		method = params[0].(string)
		if s.methods[method] != nil {
			return true
		} else {
			return false
		}
	})
}
