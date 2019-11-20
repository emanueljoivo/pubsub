package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

const (
	SubPort     = "9000"
	NetworkType = "tcp"
)

var (
	ConnectionErr = errors.New("connection error")
	UnmarshalErr  = errors.New("unmarshal error")
)

type TopicMessage struct {
	Topic   string
	Message string
}

func consume(c net.Conn) {
	for {
		message, err := bufio.NewReader(c).ReadBytes('\n')
		if err != nil {
			log.Fatalln(ConnectionErr)
		}

		var msg TopicMessage
		err = json.Unmarshal(message, &msg)

		if err != nil {
			log.Println(UnmarshalErr)
		}

		log.Printf(msg.Message)
	}
}

func NewConsumer() {
	conn, err := net.Dial(NetworkType, ":"+SubPort)

	for err != nil {
		conn, err = net.Dial(NetworkType, ":"+SubPort)
		log.Println("Waiting connection with broker")
		time.Sleep(10 * time.Second)
	}

	defer conn.Close()

	log.Println("Initializing consumer on port :" + SubPort)
	consume(conn)
}

func main() {
	go NewConsumer()
	go NewConsumer()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	<-c
}
