package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var listenURI string
var cloudURI string
var apiKey string

func main() {
	flag.StringVar(&listenURI, "listen-uri", "0.0.0.0:8181", "Local TCP server address")
	flag.StringVar(&cloudURI, "cloud-uri", "", "Cloud URI that edge-proxy service")
	flag.StringVar(&apiKey, "api-key", "", "API Key that needs to connect to edge-proxy service")
	flag.Parse()

	if cloudURI == "" {
		fmt.Printf("cloud-uri must be provided\n")

		os.Exit(1)
	}

	if apiKey == "" {
		fmt.Printf("api-key must be provided\n")

		os.Exit(1)
	}

	// setup tcp server
	listener, err := net.Listen("tcp", listenURI)
	if err != nil {
		fmt.Printf("Error listening: %s\n", err.Error())

		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Listening on: %s\n", listenURI)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()

		for {
			tcpConn, err := listener.Accept()
			if err != nil {
				fmt.Printf("Error accepting: %s\n", err.Error())

				return
			}

			fmt.Printf("Received message from %s to %s\n", tcpConn.RemoteAddr(), tcpConn.LocalAddr())

			wsConn, _, err := websocket.DefaultDialer.Dial(cloudURI, http.Header{
				"Authorization": []string{
					"Bearer " + apiKey,
				},
			})

			if err != nil {
				fmt.Printf("Error dial into edge-proxy service: %s\n, cloudURI - %s\n", err.Error(), cloudURI)

				return
			}

			ctx, cancel := context.WithCancel(ctx)

			go func() {
				for {
					bytes := make([]byte, 4096)
					length, err := tcpConn.Read(bytes)
					if err != nil {
						fmt.Printf("failed to read bytes from tcp connection: %s\n", err.Error())

						break
					}

					if err := wsConn.WriteMessage(websocket.BinaryMessage, bytes[:length]); err != nil {
						fmt.Printf("failed to write binary message to websocket connection: %s\n", err.Error())

						break
					}
				}

				cancel()
			}()

			go func() {
				for {
					_, data, err := wsConn.ReadMessage()
					if err != nil {
						fmt.Printf("failed to read binary message from websocket connection: %s\n", err.Error())

						break
					}

					_, err = tcpConn.Write(data)
					if err != nil {
						fmt.Printf("failed to write bytes to tcp connection: %s\n", err.Error())

						break
					}
				}

				cancel()
			}()

			<-ctx.Done()
		}
	}()

	<-ctx.Done()
}

