package RPC

import (
	"encoding/json"
	"log"
	"net"
	"time"
)

type Center struct {
	Addr        string
	Service     map[string][]interface{}
	IsAvailable map[string]time.Time
	AliveTime   time.Duration
	Timeout     time.Duration
}

func NewCenter(addr string, timeout time.Duration, aliveTime time.Duration) *Center {
	center := Center{
		Addr:        addr,
		Service:     make(map[string][]interface{}), //key为addr
		IsAvailable: make(map[string]time.Time),     //某个addr的server是否有效
		Timeout:     timeout,
		AliveTime:   aliveTime,
	}
	return &center
}

func RegisterToCenter(center Center, method string, addr string) {
	center.Service[addr] = append(center.Service[addr], method)
	center.IsAvailable[addr] = time.Now().Add(center.AliveTime)
}

func (c *Center) ServiceFound(method string) []string {
	var result []string
	for addr, server := range c.Service {
		if c.IsAvailable[addr].Before(time.Now()) {
			continue
		}
		for _, m := range server {
			if method == m {
				result = append(result, addr)
			}
		}
	}
	return result
}

func (c *Center) KeepAlive(addr string) {
	c.IsAvailable[addr] = time.Now().Add(c.AliveTime)
}

func (c *Center) Serve(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("conn close error")
			return
		}
	}(conn)

	//设置读取超时时间
	if err := conn.SetReadDeadline(time.Now().Add(c.Timeout)); err != nil {
		return
	}

	var message Message
	message = Read(conn)
	//log.Println(message)

	//服务注册
	log.Println(message)
	if message.Type == 3 {
		r := message.Payload.(map[string]interface{})
		method := r["methods"].(string)
		addr := r["addr"].(string)

		RegisterToCenter(*c, method, addr)
	}

	//keep alive
	if message.Type == 2 {
		addr := message.Payload.(string)
		c.KeepAlive(addr)
	}

	//服务发现
	var request Request
	if message.Type == 4 {
		r := message.Payload.(map[string]interface{})
		request.Method = r["method"].(string)
		if r["params"] == nil {
			request.Params = nil
		} else {
			request.Params = r["params"].([]interface{})
		}

		addr := c.ServiceFound(request.Method)

		message.Type = 5
		message.Payload = addr
		l, _ := json.Marshal(addr)
		message.Length = uint16(len(l))
		Write(conn, message)
	}
}
