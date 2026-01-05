package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")

	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	defer file.Close()

	buffer := make([]byte, 8)
	currLine := ""

	for {
		n, err := file.Read(buffer)

		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Println("Error reading file:", err)
			return
		}

		text := string(buffer[:n])
		lines := strings.Split(text, "\n")

		for i, line := range lines {
			currLine += line

			if i < len(lines)-1 {
				fmt.Printf("read: %s\n", currLine)
				currLine = ""
			}
		}
	}
}
