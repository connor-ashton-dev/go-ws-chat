package client

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"golang.org/x/net/websocket"
)

type Client struct {
	port int
	name string
}

func New(p int, n string) *Client {
	return &Client{
		port: p,
		name: n,
	}
}

const (
	colorBlue = "\033[34m"
)

func (c *Client) Start() {
	host := fmt.Sprintf("localhost:%d", c.port)
	wsAddress := fmt.Sprintf("ws://%s/ws", host)
	httpAddress := fmt.Sprintf("http://%s/", host)
	ws, err := websocket.Dial(wsAddress, "", httpAddress)
	if err != nil {
		log.Fatalf("Error connecting to: %s \n\n ERROR: %v", wsAddress, err)
	}
	fmt.Println("Connected to server successfully")
	incomingMessages := make(chan string)

	go readClientMessages(ws, incomingMessages)
	go sendMessages(c, ws)

	for msg := range incomingMessages {
		fmt.Printf("\r\033[K") // Clear the current line

		if strings.HasPrefix(msg, c.name+":") {
			// If the message is from the client, print it in blue
			fmt.Println(colorBlue, msg, "\033[0m")
		} else {
			// Otherwise, print it normally
			fmt.Println(msg)
		}

		fmt.Print(colorBlue, c.name+": ", "\033[0m") // Re-print the input prompt in blue
	}
}

func sendMessages(c *Client, ws *websocket.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(colorBlue, c.name+": ", "\033[0m") // Print the prompt in blue
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Read error on client:", err)
			continue
		}

		message = strings.TrimSuffix(message, "\n")

		// Send only the plain text message
		_, err = ws.Write([]byte(c.name + ": " + message))
		if err != nil {
			fmt.Println("Write error on client:", err)
		}
	}
}

func readClientMessages(ws *websocket.Conn, incomingMessages chan string) {
	for {
		var message string
		err := websocket.Message.Receive(ws, &message)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed")
				close(incomingMessages)
				break
			}

			fmt.Println("Error recieving message: ", err)
			continue

		}
		incomingMessages <- message
	}

}
