package main

import (
	"RPC"
	"log"
	"sync"
	"time"
)

func main() {
	client := RPC.NewClient(1 * time.Second)

	network := "tcp"
	addr := "localhost:8888"

	exist, err := client.IsExistService(network, addr, "Add")
	if err != nil {
		log.Println("check method error: " + err.Error())
		return
	}
	if exist {
		log.Println("method exist")
	} else {
		log.Println("method don't exist")
		return
	}

	var a interface{} = 1
	var b interface{} = 2
	result, err := client.Call(network, addr, "Add", a, b)
	if err != nil {
		log.Println("call error:" + err.Error())
		return
	}
	log.Println("Result:", result)

	var c interface{} = 3
	var d interface{} = 4
	result, err = client.Call(network, addr, "Add", c, d)
	if err != nil {
		log.Println("call error:" + err.Error())
		return
	}
	log.Println("Result:", result)

	_, err = client.Call(network, addr, "Loop")
	log.Println(err)

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			result, _ = client.Call(network, addr, "Add", 1, 1)
			log.Println("Result: ", result)
			wg.Done()
		}()
	}
	wg.Wait()
}
