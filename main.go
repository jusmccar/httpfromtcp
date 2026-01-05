package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	f, err := os.Open("messages.txt")

	if err != nil {
		log.Fatal("Error opening file:", err)
	}

	lines := getLinesChannel(f)

	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(out)

		str := ""

		for {
			data := make([]byte, 8)
			n, err := f.Read(data)

			if err != nil {
				if err == io.EOF {
					break
				}

				log.Fatal("Error reading file:", err)
			}

			text := string(data[:n])
			lines := strings.Split(text, "\n")

			for i, line := range lines {
				str += line

				if i < len(lines)-1 {
					out <- str
					str = ""
				}
			}
		}
	}()

	return out
}
