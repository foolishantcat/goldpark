package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

func Run(remote_addr string) {
	client, err := net.DialTimeout("tcp", remote_addr, 10*time.Second)
	if err != nil {
		fmt.Printf("Connect Failed!ErrString=%s\n", err.Error())
		return
	}

	for {
		data := "hello"
		var send_buf []byte

		send_buf = []byte(data)
		fmt.Println("data=", send_buf)
		size, err := client.Write(send_buf)
		if err != nil {
			fmt.Printf("Write Failed!ErrString=%s\n", err.Error())
			return
		}

		buf := make([]byte, 256)
		size, err = client.Read(buf)
		if err != nil {
			fmt.Printf("Read Failed!ErrString=%s\n", err.Error())
			return
		}

		fmt.Printf("Size=%d,Recv Data=%s\n", size, string(buf[:size]))
	}
}

func main() {
	args := os.Args
	if len(args) < 3 {
		fmt.Printf("Params Not Right!\n")
		return
	}

	var addr string

	counts, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("Params Format Not Right\n")
		return
	}

	addr = args[1]
	for i := 0; i < counts; i++ {

		go Run(addr)
	}

	for {
		time.Sleep(2 * time.Second)
		fmt.Printf("sleep!\n")
	}
}
