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
var applicationID string

func main() {
	flag.StringVar(&listenURI, "listen-uri", "0.0.0.0:8181", "Local TCP server address")
	flag.StringVar(&cloudURI, "cloud-uri", "", "Cloud URI that edge tunneling service is running on")
	flag.StringVar(&apiKey, "api-key", "", "API Key or JWT Token that needs to connect to edge tunneling service")
	flag.StringVar(&applicationID, "application-id", "", "If JWT is used instead of api-key, application-id is needed")
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
				"X-Application-ID": []string{applicationID},
			})

			if err != nil {
				fmt.Printf("Error dial into edge tunneling service: %s\n, cloudURI - %s\n", err.Error(), cloudURI)

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

					fmt.Printf("write %d bytes of data to websocket connection\n", length)
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

					length, err := tcpConn.Write(data)
					if err != nil {
						fmt.Printf("failed to write bytes to tcp connection: %s\n", err.Error())

						break
					}

					fmt.Printf("write %d bytes of data back to local tcp connection\n", length)
				}

				cancel()
			}()

			<-ctx.Done()
		}
	}()

	<-ctx.Done()
}
