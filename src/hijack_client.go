package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	// Send Upgrade request
	fmt.Fprintf(conn, "GET /api/stream HTTP/1.1\r\n")
	fmt.Fprintf(conn, "Host: localhost:8080\r\n")
	fmt.Fprintf(conn, "Connection: Upgrade\r\n")
	fmt.Fprintf(conn, "Upgrade: testproto\r\n")
	fmt.Fprintf(conn, "\r\n")

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		fmt.Fprintf(conn, "%s\n", input.Text())
	}
}
