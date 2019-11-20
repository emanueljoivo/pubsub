package main

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"time"
)

type TopicMessage struct {
	Topic   string
	Message string
}

const (
	PubPort       = "8000"
	NetworkType   = "tcp"
	charset       = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	ConnectionErr = errors.New("connection error")
)

func stringWithCharset(length int, charset string) string {
	var seededRand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func generateString(length int) string {
	return stringWithCharset(length, charset)
}

func produce(c net.Conn, topic string) {
	for {
		message := &TopicMessage{
			Topic:   topic,
			Message: generateString(10),
		}

		e, err := json.Marshal(message)

		if err != nil {
			log.Println(ConnectionErr)
		}

		_, err = c.Write(append(e, '\n'))

		if err != nil {
			log.Fatalln(ConnectionErr)
		}
		log.Println(message.Message)
		time.Sleep(2 * time.Second)
	}
}

func NewProducer() {
	conn, err := net.Dial(NetworkType, ":"+PubPort)

	for err != nil {
		conn, err = net.Dial(NetworkType, ":"+PubPort)
		log.Println("Waiting connection with broker")
		time.Sleep(10 * time.Second)
	}

	defer conn.Close()

	log.Println("Initializing producer on port :" + PubPort)

	rand.Seed(time.Now().Unix())
	produce(conn, "topic")
}

func main() {
	go NewProducer()
	go NewProducer()
	go NewProducer()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	<-c
}
