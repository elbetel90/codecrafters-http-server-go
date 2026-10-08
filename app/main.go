package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/app/types"
)

var (
	READ_BUF_SIZE int = 1024
)

func handleClient(conn net.Conn, directory string) {
	defer conn.Close()
	buf := make([]byte, READ_BUF_SIZE)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		return
	}
	request := string(buf[:n])
	request_line, _, _ := strings.Cut(request, "\r\n")
	parts := strings.Split(request_line, " ")
	response := "HTTP/1.1 404 Not Found\r\n\r\n"
	if len(parts) >= 2 {
		method := parts[0]
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
		} else if file_name, found := strings.CutPrefix(path, "/files/"); found {
			full_path := filepath.Join(directory, file_name)
			if method == string(types.HttpMethodPOST) {
				_, body, _ := strings.Cut(request, "\r\n\r\n")
				content_length := 0
				for _, header := range strings.Split(request, "\r\n")[1:] {
					name, value, ok := strings.Cut(header, ":")
					if !ok {
						continue
					}
					if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
						content_length, _ = strconv.Atoi(strings.TrimSpace(value))
						break
					}
					if content_length > len(body) {
						content_length = len(body)
					}
				}
				err = os.WriteFile(full_path, []byte(body[:content_length]), 0644)
				if err == nil {
					response = "HTTP/1.1 201 Created\r\n\r\n"
				}
			} else {
				content, err := os.ReadFile(full_path)
				if err == nil {
					response = fmt.Sprintf(
						"HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s",
						len(content),
						content,
					)
				}
			}
		}
	}
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
	}
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	directory := flag.String("directory", "", "path")
	flag.Parse()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleClient(conn, *directory)

	}
}
