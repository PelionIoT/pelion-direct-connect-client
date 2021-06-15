package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var listenURI string
var cloudURI string
var apiKey string
var applicationID string
var sendPings bool
func main() {
	flag.StringVar(&listenURI, "listen-uri", "0.0.0.0:8181", "Local TCP server address")
	flag.StringVar(&cloudURI, "cloud-uri", "", "Cloud URI that edge tunneling service is running on")
	flag.StringVar(&apiKey, "api-key", "", "API Key or JWT Token that needs to connect to edge tunneling service")
	flag.StringVar(&applicationID, "application-id", "", "If JWT is used instead of api-key, application-id is needed")
	flag.BoolVar(&sendPings, "send-pings", true, "Keep the tunnel alive by pinging the other side of the tunnel")
	flag.Parse()

	customFormatter := new(logrus.TextFormatter)
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
	logrus.SetLevel(logrus.DebugLevel)

	if cloudURI == "" {
		logrus.Errorf("cloud-uri must be provided\n")

		os.Exit(1)
	}

	if apiKey == "" {
		logrus.Errorf("api-key must be provided\n")

		os.Exit(1)
	}

	// setup tcp server
	listener, err := net.Listen("tcp", listenURI)
	if err != nil {
		logrus.Errorf("Error listening: %s\n", err.Error())

		os.Exit(1)
	}
	defer listener.Close()

	logrus.Infof("Listening on: %s\n", listenURI)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()

		for {
			tcpConn, err := listener.Accept()
			if err != nil {
				logrus.Errorf("Error accepting: %s\n", err.Error())

				return
			}

			logrus.Debugf("Received message from %s to %s\n", tcpConn.RemoteAddr(), tcpConn.LocalAddr())

			wsConn, _, err := websocket.DefaultDialer.Dial(cloudURI, http.Header{
				"Authorization": []string{
					"Bearer " + apiKey,
				},
				"X-Application-ID": []string{applicationID},
			})

			if err != nil {
				logrus.Errorf("Error dial into edge tunneling service: %s\n, cloudURI - %s\n", err.Error(), cloudURI)

				return
			}

			ctx, cancel := context.WithCancel(ctx)

			go func() {
				chanLength := make(chan int)
				bytes := make([]byte, 4096)
				go func () {
					for{
						length, err := tcpConn.Read(bytes)
						if err != nil {
							logrus.Errorf("failed to read bytes from tcp connection: %s\n", err.Error())
							cancel()
							return
						}
						chanLength <- length
					}
				}()

				t := time.NewTicker(time.Duration(5 * time.Second))
				if sendPings {
					defer t.Stop()
					logrus.Infof("Send pings enabled\n")
				} else {
					logrus.Infof("Send pings disabled\n")
					t.Stop()
				}

				for {
					select {
					case length := <- chanLength:
							if err := wsConn.WriteMessage(websocket.BinaryMessage, bytes[:length]); err != nil {
								logrus.Errorf("failed to write binary message to websocket connection: %s\n", err.Error())
							cancel()
							return
						}

						logrus.Debugf("write %d bytes of data to websocket connection\n", length)

					case <-t.C:
						if err := wsConn.WriteControl(websocket.PingMessage, []byte(""), time.Now().Add(time.Second)); err != nil {
							logrus.Errorf("Error writing ping %#v\n", err)
							cancel()
							return
						}
						logrus.Debugf("Wrote Ping\n")
					}
				}
			}()

			go func() {
				for {
					_, data, err := wsConn.ReadMessage()
					if err != nil {
						logrus.Errorf("failed to read binary message from websocket connection: %s\n", err.Error())

						break
					}

					length, err := tcpConn.Write(data)
					if err != nil {
						logrus.Errorf("failed to write bytes to tcp connection: %s\n", err.Error())

						break
					}

					logrus.Debugf("write %d bytes of data back to local tcp connection\n", length)
				}

				cancel()
			}()

			<-ctx.Done()
		}
	}()

	<-ctx.Done()
}
