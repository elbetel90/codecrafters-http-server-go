package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

var (
	READ_BUF_SIZE int = 1024
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
		buf := make([]byte, READ_BUF_SIZE)
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
		if len(parts) >= 2 {
			path := parts[1]
			if path == "/" {
				response = "HTTP/1.1 200 OK\r\n\r\n"
			} else if strings.HasPrefix(path, "/echo") {
				str := strings.TrimPrefix(path, "/echo/")
				response = fmt.Sprintf(
					"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
					len(str),
					str,
				)
			} else if strings.HasPrefix(path, "/user-agent") {
				request_parts := strings.Split(request, "\r\n")
				headers := request_parts[1:]
				for _, header := range headers {
					name, value, found := strings.Cut(header, ":")
					if !found {
						continue
					}
					if strings.EqualFold(strings.TrimSpace(name), "User-Agent") {
						user_agent_header_value := strings.TrimSpace(value)
						response = fmt.Sprintf(
							"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
							len(user_agent_header_value),
							user_agent_header_value,
						)
					}
				}
			}
		}
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing response: ", err.Error())
		}
		conn.Close()
	}
}
