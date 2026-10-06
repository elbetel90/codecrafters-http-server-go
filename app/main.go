package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading request: ", err.Error())
			conn.Close()
			continue
		}
		request := string(buf[:n])
		request_line, _, _ := strings.Cut(request, "\r\n")
		parts := strings.Split(request_line, " ")
		response := "HTTP/1.1 404 Not Found\r\n\r\n"
		if len(parts) >= 2 && parts[1] == "/" {
			response = "HTTP/1.1 200 OK\r\n\r\n"
		}

		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing response: ", err.Error())
		}
		conn.Close()
	}
}
