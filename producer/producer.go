package main

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
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

func handleConnection(c net.Conn, topic string) {
	for {
		message := &TopicMessage{
			Topic:   topic,
			Message: generateString(10),
		}

		e, err := json.Marshal(message)

		if err != nil {
			log.Println(ConnectionErr)
		}

		bytesWrited, err = c.Write(append(e, '\n'))

		log.Println("Bytes written ")

		if err != nil {
			log.Fatalln(ConnectionErr)
		}
		log.Println(message.Message)
		time.Sleep(2 * time.Second)
	}
}

func NewProducer(nTopics int) {
	conn, err := net.Dial(NetworkType, ":"+PubPort)

	for err != nil {
		conn, err = net.Dial(NetworkType, ":"+PubPort)
		log.Println("Waiting connection with broker")
		time.Sleep(10 * time.Second)
	}

	defer conn.Close()

	log.Println("Initializing producer on port :" + PubPort)

	rand.Seed(time.Now().Unix())
	for i := 0; i < nTopics; i++ {
		go handleConnection(conn, "topic_" + strconv.Itoa(i))
	}
}

func main() {
	var nProducers = 1
	var nTopics = 1

	for i := 0; i < nProducers; i++ {
		go NewProducer(nTopics)
	}

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	<-c

}
