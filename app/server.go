package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	port := flag.String("port", "6379", "Port to run the server on")

	flag.Parse()

	address := "0.0.0.0:" + *port

	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Failed to bind to port " + *port)
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
		resp := NewResp(conn)
		writer := NewWriter(conn)
		value, err := resp.read()
		if err != nil {
			fmt.Println(err)
			return
		}
		if value.typ != "array" {
			fmt.Println("Invalid request, expected array")
			continue
		}
		if len(value.array) == 0 {
			fmt.Println("Invalid request, expected array length > 0")
			continue
		}
		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]
		handler, ok := Handlers[command]
		if !ok {
			fmt.Println("Invalid command: ", command)
			writer.Write(Value{typ: "string", str: ""})
			continue
		}
		response := handler(args)
		writer.Write(response)
	}
}
