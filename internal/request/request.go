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

const crlf = "\r\n"
const bufferSize = 8

var ErrorMalformedRequestLine = fmt.Errorf("Malformed request line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("Unsupported HTTP version")
var ErrorTryingToReadDataInADoneState = fmt.Errorf("Trying to read data in a done state")
var ErrorUnknownState = fmt.Errorf("Unknown state")

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{state: StateInit}
	data := make([]byte, bufferSize)
	readToIndex := 0

	for request.state != StateDone {
		if readToIndex == len(data) {
			tempData := make([]byte, len(data)*2)
			copy(tempData, data)
			data = tempData
		}

		bytesRead, err := reader.Read(data[readToIndex:])

		if err != nil {
			return nil, err
		}

		readToIndex += bytesRead
		bytesConsumed, err := request.Parse(data[:readToIndex])

		if err != nil {
			return nil, err
		}

		copy(data, data[bytesConsumed:readToIndex])
		readToIndex -= bytesConsumed
	}

	return request, nil
}

func (r *Request) Parse(data []byte) (int, error) {
	if r.state == StateDone {
		return 0, ErrorTryingToReadDataInADoneState
	}

	if r.state != StateInit {
		return 0, ErrorUnknownState
	}

	requestLine, bytesConsumed, err := parseRequestLine(string(data[:]))

	if err != nil {
		return 0, err
	}

	if bytesConsumed == 0 {
		return 0, nil
	}

	r.RequestLine = *requestLine
	r.state = StateDone

	return bytesConsumed, nil
}

func parseRequestLine(str string) (*RequestLine, int, error) {
	if !strings.Contains(str, crlf) {
		return nil, 0, nil
	}

	lines := strings.Split(str, crlf)
	bytesConsumed := len(lines[0]) + len(crlf)

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
