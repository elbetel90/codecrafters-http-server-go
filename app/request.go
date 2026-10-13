package main

import (
	"errors"
	"strconv"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    []byte
}

func (r Request) Header(name string) string {
	if r.Headers == nil {
		return ""
	}
	return r.Headers[strings.ToLower(name)]
}

func parseRequest(raw string) (Request, error) {
	head, body, _ := strings.Cut(raw, "\r\n\r\n")
	lines := strings.Split(head, "\r\n")
	if len(lines) == 0 || lines[0] == "" {
		return Request{}, errors.New("empty request")
	}
	parts := strings.Split(lines[0], " ")
	if len(parts) < 2 {
		return Request{}, errors.New("malformed request line")
	}
	req := Request{
		Method:  parts[0],
		Path:    parts[1],
		Headers: make(map[string]string),
		Body:    []byte(body),
	}

	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(name))
		req.Headers[key] = strings.TrimSpace(value)
	}

	if cl := req.Header("Content-Length"); cl != "" {
		n, err := strconv.Atoi(cl)
		if err == nil {
			if n < 0 {
				n = 0
			}
			if n > len(req.Body) {
				n = len(req.Body)
			}
			req.Body = req.Body[:n]
		}
	}
	return req, nil
}
