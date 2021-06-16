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
var pingDuration int
func main() {
	flag.StringVar(&listenURI, "listen-uri", "0.0.0.0:8181", "Local TCP server address")
	flag.StringVar(&cloudURI, "cloud-uri", "", "Cloud URI that edge tunneling service is running on")
	flag.StringVar(&apiKey, "api-key", "", "API Key or JWT Token that needs to connect to edge tunneling service")
	flag.StringVar(&applicationID, "application-id", "", "If JWT is used instead of api-key, application-id is needed")
	flag.IntVar(&pingDuration, "ping-duration", 5, "Ping duration in seconds, 0 will disable pings to the tunnel")
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

	if pingDuration > 0 {
		logrus.Infof("Send pings enabled with an interval %v seconds\n", pingDuration)
	} else {
		logrus.Infof("Send pings disabled\n")
	}

	logrus.Infof("Listening on: %s\n", listenURI)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			logrus.Errorf("Error accepting: %s\n", err.Error())

			return
		}

		logrus.Debugf("(Connection from %s to %s) Starting\n", tcpConn.RemoteAddr(), tcpConn.LocalAddr())

		wsConn, _, err := websocket.DefaultDialer.Dial(cloudURI, http.Header{
			"Authorization": []string{
				"Bearer " + apiKey,
			},
			"X-Application-ID": []string{applicationID},
		})

		if err != nil {
			logrus.Errorf("(Connection from %s to %s) Error dial into edge tunneling service: %s\n, cloudURI - %s\n", tcpConn.RemoteAddr(), tcpConn.LocalAddr(), err.Error(), cloudURI)

			return
		}

		_, cancel := context.WithCancel(ctx)

		go func() {
			chanBytes := make(chan []byte, 4096)
			go func () {
				for{
					bytes := make([]byte, 4096)
					length, err := tcpConn.Read(bytes)
					if err != nil {
						logrus.Errorf("(Connection from %s to %s) failed to read bytes from tcp connection: %s\n", tcpConn.RemoteAddr(), tcpConn.LocalAddr(), err.Error())
						cancel()

						return
					}
					chanBytes <- bytes[:length]
				}
			}()

			c := time.Tick(time.Duration(pingDuration) * time.Second)
			logrus.Infof("pingDuration :%v c:%#v\n", pingDuration, c)

			for {
				select {
				case bytes := <- chanBytes:
					length := len(bytes)
					if err := wsConn.WriteMessage(websocket.BinaryMessage, bytes[:length]); err != nil {
							logrus.Errorf("(Connection from %s to %s) failed to write binary message to websocket connection: %s\n", tcpConn.RemoteAddr(), tcpConn.LocalAddr(), err.Error())
						cancel()

						return
					}

					logrus.Debugf("write %d bytes of data to websocket connection\n", length)

				case <-c:
					if err := wsConn.WriteControl(websocket.PingMessage, []byte(""), time.Now().Add(time.Second)); err != nil {
						logrus.Errorf("(Connection from %s to %s) Error writing ping %s\n", tcpConn.RemoteAddr(), tcpConn.LocalAddr(), err)
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
					logrus.Errorf("(Connection from %s to %s) failed to read binary message from websocket connection: %s\n", tcpConn.RemoteAddr(), tcpConn.LocalAddr(), err.Error())

					break
				}

				length, err := tcpConn.Write(data)
				if err != nil {
					logrus.Errorf("(Connection from %s to %s) failed to write bytes to tcp connection: %s\n", tcpConn.RemoteAddr(), tcpConn.LocalAddr(), err.Error())

					break
				}

				logrus.Debugf("write %d bytes of data back to local tcp connection\n", length)
			}

			cancel()
		}()

		logrus.Infof("(Connection from %s to %s) Dispatched\n", tcpConn.RemoteAddr(), tcpConn.LocalAddr())
	}

}
