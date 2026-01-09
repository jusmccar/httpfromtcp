package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"httpfromtcp/internal/headers"
)

type ParserState int

const (
	requestStateInit ParserState = iota
	requestStateDone
	requestStateParsingHeaders
	requestStateParsingBody
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        Body
	state       ParserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Body []byte

const crlf = "\r\n"
const bufferSize = 8

var (
	ErrorMalformedRequestLine   = fmt.Errorf("Malformed request line")
	ErrorUnsupportedHttpVersion = fmt.Errorf("Unsupported HTTP version")
	ErrorMalformedBody          = fmt.Errorf("Malformed body")
	ErrorUnknownState           = fmt.Errorf("Unknown state")
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := &Request{
		state:   requestStateInit,
		Headers: headers.NewHeaders(),
	}

	data := make([]byte, bufferSize)
	readToIndex := 0

	for r.state != requestStateDone {
		if readToIndex == len(data) {
			tempData := make([]byte, len(data)*2)
			copy(tempData, data)
			data = tempData
		}

		bytesRead, err := reader.Read(data[readToIndex:])

		if err != nil {
			if err != io.EOF || readToIndex == 0 {
				return nil, err
			}
		}

		readToIndex += bytesRead
		bytesConsumed, err := r.parse(data[:readToIndex])

		if err != nil {
			return nil, err
		}

		copy(data, data[bytesConsumed:readToIndex])
		readToIndex -= bytesConsumed
	}

	return r, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])

		if err != nil {
			return 0, err
		}

		if n == 0 {
			break
		}

		totalBytesParsed += n
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInit:
		requestLine, bytesConsumed, err := parseRequestLine(string(data))

		if err != nil {
			return 0, err
		}

		if bytesConsumed == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return bytesConsumed, nil

	case requestStateParsingHeaders:
		bytesConsumed, done, err := r.Headers.Parse(data)

		if err != nil {
			return 0, err
		}

		if done {
			r.state = requestStateParsingBody

			// parse final crlf
			return 2, nil
		}

		return bytesConsumed, nil

	case requestStateParsingBody:
		contentLengthStr := r.Headers.Get("Content-Length")

		if contentLengthStr == "" {
			r.state = requestStateDone

			return 0, nil
		}

		r.Body = append(r.Body, data...)

		contentLength, err := strconv.Atoi(contentLengthStr)

		if err != nil {
			return 0, err
		}

		if len(r.Body) > contentLength {
			return 0, ErrorMalformedBody
		} else if len(r.Body) == contentLength {
			r.state = requestStateDone
		}

		return len(data), nil

	case requestStateDone:
		return 0, nil

	default:
		return 0, ErrorUnknownState
	}
}

func parseRequestLine(str string) (*RequestLine, int, error) {
	if !strings.Contains(str, crlf) {
		return nil, 0, nil
	}

	lines := strings.Split(str, crlf)
	request := lines[0]
	bytesConsumed := len(request) + len(crlf)
	parts := strings.Split(request, " ")

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

	requestLine := &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   httpVersion,
	}

	return requestLine, bytesConsumed, nil
}
