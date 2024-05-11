package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		var response string = "+PONG\r\n"
		buffer := make([]byte, 1024)
		data, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading command: ", err.Error())
		}
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error sending response:", err.Error())
			return
		}
		fmt.Println("Response sent to client: ", response)
	}
}
