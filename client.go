package RPC

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"time"
)

type Client struct {
	Timeout time.Duration
}

func NewClient(timeout time.Duration) *Client {
	return &Client{
		Timeout: timeout,
	}
}

func (c *Client) Call(network string, addr string, method string, params ...interface{}) (interface{}, error) {
	var message Message
	conn, err := net.DialTimeout(network, addr, c.Timeout)
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

	var request interface{} = Request{
		Method: method,
		Params: params,
	}
	message.Type = 4
	r, _ := json.Marshal(request)
	message.Length = uint16(len(r))
	message.Payload = request

	//设置写入超时时间
	if err := conn.SetWriteDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, err
	}

	message = Write(conn, message)
	if message.Type == 0 {
		log.Printf("client write request error")
		return nil, errors.New("client write request error")
	}

	// 设置读取超时时间
	if err := conn.SetReadDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, err
	}

	message = Read(conn)
	if message.Type == 0 {
		return nil, errors.New(message.Payload.(string))
	} else {
		var response Response
		resp := message.Payload.(map[string]interface{})
		response.Error = resp["error"].(string)
		response.Result = resp["result"].(interface{})
		if response.Error != "" {
			return nil, errors.New(response.Error)
		} else {
			return response.Result, nil
		}
	}
}

func (c *Client) IsExistService(network string, addr string, method string) (bool, error) {
	exist, err := c.Call(network, addr, "CheckMethod", method)
	if err != nil {
		log.Printf(err.Error())
		return false, err
	} else {
		return exist.(bool), err
	}
}

func (c *Client) Find(addr string, method string, params ...interface{}) ([]string, error) {
	var message Message

	conn, err := net.DialTimeout("tcp", addr, c.Timeout)
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

	var request interface{} = Request{
		Method: method,
		Params: params,
	}
	message.Type = 4
	r, _ := json.Marshal(request)
	message.Length = uint16(len(r))
	message.Payload = request

	//设置写入超时时间
	if err := conn.SetWriteDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, err
	}

	message = Write(conn, message)
	if message.Type == 0 {
		log.Printf("client write request error")
		return nil, errors.New("client write request error")
	}

	// 设置读取超时时间
	if err := conn.SetReadDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, err
	}

	message = Read(conn)
	if message.Type == 0 {
		return nil, errors.New(message.Payload.(string))
	} else {
		resp := message.Payload
		if resp == nil {
			return nil, nil
		} else {
			var result []string
			if response, ok := resp.([]interface{}); ok {
				for _, item := range response {
					if str, ok := item.(string); ok {
						result = append(result, str)
					}
				}
			}
			return result, nil
		}
	}
}
