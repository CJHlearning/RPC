package main

import (
	"RPC"
	"flag"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

func main() {
	client := RPC.NewClient(10 * time.Second)
	var centerAddr string
	network := "tcp"
	//centerAddr := "localhost:9999"

	centerIP := flag.String("i", "", "注册中心的IP地址(必须输入)")
	port := flag.String("p", "", "注册中心的端口(必须输入)")

	flag.Parse()

	if *centerIP == "" || *port == "" {
		flag.PrintDefaults()
		os.Exit(1)
	} else {
		centerAddr = *centerIP + ":" + *port
	}
	rand.NewSource(time.Now().UnixNano())

	var a interface{} = 1
	var b interface{} = 2
	var c interface{} = 3
	var d interface{} = 4

	addr, err := client.Find(centerAddr, "Add", a, b)
	log.Println(addr)
	if addr == nil {
		log.Println("server don't exist")
	} else {
		serverAddr := addr[rand.Intn(len(addr))]
		result, err := client.Call(network, serverAddr, "Add", a, b)
		if err != nil {
			log.Println("call error:" + err.Error())
			return
		}
		log.Println("Result:", result)
	}

	addr, err = client.Find(centerAddr, "Sub", c, d)
	if addr == nil {
		log.Println("server don't exist")
	} else {
		serverAddr := addr[rand.Intn(len(addr))]
		result, err := client.Call(network, serverAddr, "Sub", c, d)
		if err != nil {
			log.Println("call error:" + err.Error())
			return
		}
		log.Println("Result:", result)
	}

	addr, err = client.Find(centerAddr, "Mul", c, d)
	if addr == nil {
		log.Println("server don't exist")
	} else {
		serverAddr := addr[rand.Intn(len(addr))]
		result, err := client.Call(network, serverAddr, "Mul", c, d)
		if err != nil {
			log.Println("call error:" + err.Error())
			return
		}
		log.Println("Result:", result)
	}

	addr, err = client.Find(centerAddr, "Div", c, d)
	if addr == nil {
		log.Println("server don't exist")
	} else {
		serverAddr := addr[rand.Intn(len(addr))]
		result, err := client.Call(network, serverAddr, "Div", c, d)
		if err != nil {
			log.Println("call error:" + err.Error())
			return
		}
		log.Println("Result:", result)
	}

	addr, err = client.Find(centerAddr, "Square", c)
	if addr == nil {
		log.Println("server don't exist")
	} else {
		serverAddr := addr[rand.Intn(len(addr))]
		result, err := client.Call(network, serverAddr, "Square", c)
		if err != nil {
			log.Println("call error:" + err.Error())
			return
		}
		log.Println("Result:", result)
	}

	addr, err = client.Find(centerAddr, "Big", c, d)
	if addr == nil {
		log.Println("server don't exist")
	} else {
		serverAddr := addr[rand.Intn(len(addr))]
		result, err := client.Call(network, serverAddr, "Big", c, d)
		if err != nil {
			log.Println("call error:" + err.Error())
			return
		}
		log.Println("Result:", result)
	}

	addr, err = client.Find(centerAddr, "Small", c, d)
	if addr == nil {
		log.Println("server don't exist")
	} else {
		serverAddr := addr[rand.Intn(len(addr))]
		result, err := client.Call(network, serverAddr, "Small", c, d)
		if err != nil {
			log.Println("call error:" + err.Error())
			return
		}
		log.Println("Result:", result)
	}

	addr, err = client.Find(centerAddr, "Equal", c, d)
	if addr == nil {
		log.Println("server don't exist")
	} else {
		serverAddr := addr[rand.Intn(len(addr))]
		result, err := client.Call(network, serverAddr, "Equal", c, d)
		if err != nil {
			log.Println("call error:" + err.Error())
			return
		}
		log.Println("Result:", result)
	}

	addr, err = client.Find(centerAddr, "HelloWorld")
	if addr == nil {
		log.Println("server don't exist")
	} else {
		serverAddr := addr[rand.Intn(len(addr))]
		result, _ := client.Call(network, serverAddr, "HelloWorld")
		log.Println(result)
	}

	addr, err = client.Find(centerAddr, "Loop")
	if addr == nil {
		log.Println("server don't exist")
	} else {
		serverAddr := addr[rand.Intn(len(addr))]
		_, err = client.Call(network, serverAddr, "Loop")
		log.Println(err)
	}

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			addr, err = client.Find(centerAddr, "Add", 1, 1)
			if addr == nil {
				log.Println("server don't exist")
			} else {
				serverAddr := addr[rand.Intn(len(addr))]
				result, _ := client.Call(network, serverAddr, "Add", 1, 1)
				log.Println("Result:", result)
				wg.Done()
			}
		}()
	}
	wg.Wait()
}
