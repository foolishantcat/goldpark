package main

import (
	"fmt"
	"net"
)

func Run(conn net.Conn) {

	defer conn.Close()
	buf := make([]byte, 256)
	for {

		size, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("Read Failed!ErrString=%s\n", err.Error())
			break
		}

		var data string

		data = string(buf[:size])
		data += "123"

		fmt.Printf("size=%d,Reponse Data=%s\n", size, data)
		size, err = conn.Write([]byte(data))
		if err != nil {
			fmt.Printf("Send Failed!ErrString=%s\n", err.Error())
			break
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":12357")
	if err != nil {
		fmt.Printf("listener failed!ErrString=%s", err.Error())
		return
	}
	fmt.Printf("Listen 12357 Ok!\n")

	defer listener.Close()

	// 开始收包
	for {
		accpet_conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept Failed!ErrString=%s", err.Error())
			return
		}

		fmt.Printf("New Connect!PeerIp=%s", accpet_conn.RemoteAddr())
		go Run(accpet_conn)
	}
}
