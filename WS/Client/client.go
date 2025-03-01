package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:2808")
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer conn.Close()

	request := "GET /ws HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n" +
		"Sec-WebSocket-Version: 13\r\n\r\n"

	conn.Write([]byte(request))

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Fatal("Error reading response:", response)
	}

	fmt.Println("Server response:", response)

	go func() {
		for {
			var message string
			fmt.Println("Enter message here:")
			fmt.Scan(&message)
			frame := append([]byte{0x81, byte(len(message))}, []byte(message)...)
			conn.Write(frame)
			time.Sleep(3 * time.Second)
		}
	}()

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("Server closed the connection")
			break
		}
		time.Sleep(2 * time.Second)
		fmt.Println("Received from server:", string(buf[:n]))
	}
}
