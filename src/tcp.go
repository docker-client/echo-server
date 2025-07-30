package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type Server struct {
	host string
	port string
}

type Client struct {
	conn net.Conn
}

type Config struct {
	Host string
	Port string
}

func New(config *Config) *Server {
	return &Server{
		host: config.Host,
		port: config.Port,
	}
}

func (server *Server) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", server.host, server.port))
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		client := &Client{
			conn: conn,
		}
		go client.handleRequest()
	}
}

func (client *Client) handleRequest() {
	reader := bufio.NewReader(client.conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			client.conn.Close()
			return
		}
		message = strings.Trim(message, "\n\r")
		fmt.Sprintf("Message incoming: %s\n", message)
		client.conn.Write([]byte(fmt.Sprintf("Message received: %s\n", message)))
		if message == "exit" {
			fmt.Println("Exiting connection...")
			client.conn.Close()
			return
		}
	}
}

func main() {
	server := New(&Config{
		Host: "localhost",
		Port: "8888",
	})
	fmt.Println("run...")
	server.Run()
	fmt.Println("...done")
}
