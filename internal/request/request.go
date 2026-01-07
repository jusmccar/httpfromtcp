package request

import (
	"fmt"
	"io"
	"strings"
)

type ParserState int

const (
	StateInit ParserState = 0
	StateDone ParserState = 1
)

type Request struct {
	RequestLine RequestLine
	state       ParserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var ErrorMalformedRequestLine = fmt.Errorf("Malformed request line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("Unsupported HTTP version")

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{}
	request.state = StateInit

	data := make([]byte, 1024)
	dataLen := 0

	for request.state != StateDone {
		bytesRead, err := reader.Read(data[dataLen:])

		if err != nil {
			return nil, err
		}

		dataLen += bytesRead
		bytesConsumed, err := request.parse(data[:dataLen+bytesRead])

		if err != nil {
			return nil, err
		}

		copy(data, data[bytesConsumed:dataLen])
		dataLen -= bytesConsumed
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesConsumed := 0

	for r.state != StateDone {
		requestLine, bytesConsumed, err := parseRequestLine(string(data[totalBytesConsumed:]))

		if err != nil {
			return 0, err
		}

		if bytesConsumed == 0 {
			break
		}

		r.RequestLine = *requestLine
		totalBytesConsumed += bytesConsumed
		r.state = StateDone
	}

	return totalBytesConsumed, nil
}

func parseRequestLine(str string) (*RequestLine, int, error) {
	if !strings.Contains(str, "\r\n") {
		return nil, 0, nil
	}

	lines := strings.Split(str, "\r\n")
	bytesConsumed := len(lines[0]) + len("\r\n")

	requestLine := lines[0]
	parts := strings.Split(requestLine, " ")

	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	method := parts[0]
	requestTarget := parts[1]
	httpParts := strings.Split(parts[2], "/")

	if len(httpParts) != 2 {
		return nil, 0, ErrorMalformedRequestLine
	}

	httpVersion := httpParts[1]

	if httpVersion != "1.1" {
		return nil, 0, ErrorUnsupportedHttpVersion
	}

	requestLineStruct := &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   httpVersion,
	}

	return requestLineStruct, bytesConsumed, nil
}
