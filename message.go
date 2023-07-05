package RPC

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
)

type Request struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

type Response struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
}

type Message struct {
	Length  uint16 //length为payload长度
	Type    uint8
	Payload any
}

//type = 0: 为err
//type = 1: non-error
//type = 2: 为keepalive
//type = 3: 为服务注册报文
//type = 4: request
//type = 5: response

func Read(conn net.Conn) Message {
	// 读取消息的 Length 字段
	var length uint16
	err := binary.Read(conn, binary.LittleEndian, &length)
	if err != nil {
		log.Println("read length error: " + err.Error())
		return Message{
			Length:  uint16(len(err.Error())),
			Type:    0,
			Payload: err.Error(),
		}
	}
	//log.Println(length)

	// 读取消息的 Type 字段
	var msgType uint8
	err = binary.Read(conn, binary.LittleEndian, &msgType)
	if err != nil {
		log.Println("read type error: " + err.Error())
		return Message{
			Length:  uint16(len(err.Error())),
			Type:    0,
			Payload: err.Error(),
		}
	}

	//log.Println(length)
	bodyBuf := make([]byte, length)
	_, err = io.ReadFull(conn, bodyBuf)

	var payload interface{}
	err = json.Unmarshal(bodyBuf, &payload)

	if err != nil {
		log.Println("read Unmarshal error: " + err.Error())
		return Message{
			Length:  uint16(len(err.Error())),
			Type:    0,
			Payload: err.Error(),
		}
	} else {
		return Message{
			Length:  length,
			Type:    msgType,
			Payload: payload,
		}
	}
}

func Write(conn net.Conn, message Message) Message {
	//log.Println(message.Payload)
	payloadBytes, err := json.Marshal(message.Payload)

	if err != nil {
		log.Println("response Marshal error:" + err.Error())
		return Message{
			Length:  uint16(len(err.Error())),
			Type:    0,
			Payload: err.Error(),
		}
	} else {
		// 设置消息的 Length 字段
		message.Length = uint16(len(payloadBytes))

		// 发送消息的 Length 字段给服务器
		err = binary.Write(conn, binary.LittleEndian, message.Length)
		if err != nil {
			log.Fatal("Length sending error:", err)
		}

		// 发送消息的 Type 字段给服务器
		err = binary.Write(conn, binary.LittleEndian, message.Type)
		if err != nil {
			log.Fatal("Type sending error:", err)
		}

		_, err = conn.Write(payloadBytes)
		if err != nil {
			log.Println("payload write error:" + err.Error())
			return Message{
				Length:  uint16(len(err.Error())),
				Type:    0,
				Payload: err.Error(),
			}
		} else {
			return Message{
				Type: 1,
			}
		}
	}
}
