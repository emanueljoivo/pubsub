package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"net"
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

func handleConnection(c net.Conn, topic string) {
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

func main() {
	l, err := net.Listen(NetworkType, ":"+SubPort)
	for err != nil {
		log.Println("waiting connection with broker")
	}
	log.Println("Initializing consumers on port " + SubPort)
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go handleConnection(c, "topic_unic")
	}
}
