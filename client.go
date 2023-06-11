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
	conn, err := net.DialTimeout(network, addr, c.Timeout)
	if err != nil {
		log.Println("dial failure " + err.Error())
	}

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println(conn.LocalAddr().String() + "conn close error" + err.Error())
		}
	}(conn)
	req := Request{method, params}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	//设置写入超时时间
	if err := conn.SetWriteDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, err
	}

	_, err = conn.Write(reqBytes)
	if err != nil {
		log.Printf("client write request timeout")
		return nil, err
	}

	// 设置读取超时时间
	if err := conn.SetReadDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, err
	}

	var respBytes []byte

	// 创建一个临时缓冲区
	buf := make([]byte, 1024)

	for {
		// 读取连接中的数据，并将其存储到临时缓冲区 buf 中
		n, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				log.Println("conn read request timeout")
			} else {
				log.Println("conn read error: " + err.Error())
			}
			return nil, err
		}

		// 将临时缓冲区中的数据追加到响应字节切片中
		respBytes = append(respBytes, buf[:n]...)

		// 检查是否已经读取完所有响应数据
		if n < len(buf) {
			break
		}
	}

	var resp Response
	err = json.Unmarshal(respBytes, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}

	return resp.Result, nil
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
