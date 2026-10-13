package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/app/types"
)

var (
	READ_BUF_SIZE int = 1024
)

func handleClient(conn net.Conn, directory string) {
	defer conn.Close()
	for {
		buf := make([]byte, READ_BUF_SIZE)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		request := string(buf[:n])
		req, err := parseRequest(request)
		if err != nil {
			continue
		}
		response := "HTTP/1.1 404 Not Found\r\n\r\n"
		close_conn := strings.EqualFold(req.Header("Connection"), "close")
		method := req.Method
		path := req.Path
		if path == "/" {
			response = "HTTP/1.1 200 OK\r\n\r\n"
		} else if strings.HasPrefix(path, "/echo") {
			str := strings.TrimPrefix(path, "/echo/")
			compression_scheme_str := req.Header("Accept-Encoding")
			valid_compression_scheme := ""
			for _, compression_scheme := range strings.Split(compression_scheme_str, ",") {
				scheme := strings.TrimSpace(compression_scheme)
				if scheme != "" && slices.Contains(types.SupportedCompressionSchemes(), scheme) {
					valid_compression_scheme = scheme
					break
				}
			}
			if valid_compression_scheme != "" {
				switch valid_compression_scheme {
				case string(types.CompressionSchemeGzip):
					compressed, err := gzipCompress([]byte(str))
					if err == nil {
						response = fmt.Sprintf(
							"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\nContent-Encoding: %s\r\n\r\n%s",
							len(compressed),
							valid_compression_scheme,
							compressed,
						)
					}
				}

			} else {
				response = fmt.Sprintf(
					"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
					len(str),
					str)
			}

		} else if strings.HasPrefix(path, "/user-agent") {
			user_agent := req.Header("User-Agent")
			response = fmt.Sprintf(
				"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s",
				len(user_agent),
				user_agent,
			)
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

		if close_conn {
			response = strings.Replace(response, "\r\n\r\n", "\r\nConnection: close\r\n\r\n", 1)
		}
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Error writing response: ", err.Error())
		}
		if close_conn {
			return
		}

	}
}

func gzipCompress(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write(body)
	if err != nil {
		gz.Close()
		return nil, err
	}
	err = gz.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
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
