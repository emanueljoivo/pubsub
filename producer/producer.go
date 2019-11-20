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
	ArgsErr = errors.New("args error")
	ProducersErr = errors.New("you need to specify how many producers will perform")
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

		_, err = c.Write(append(e, '\n'))

		if err != nil {
			log.Fatalln(ConnectionErr)
		}
		log.Println(message.Message)
		time.Sleep(2 * time.Second)
	}
}

func produce(nTopics int) {
	conn, err := net.Dial(NetworkType, ":"+PubPort)

	for err != nil {
		log.Println("Waiting connection with broker")
		time.Sleep(10 * time.Second)
	}

	defer conn.Close()

	log.Println("Initializing producers on port " + PubPort)

	rand.Seed(time.Now().Unix())
	for i := 0; i < nTopics; i++ {
		go handleConnection(conn, "topic_" + strconv.Itoa(i))
	}
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalln(ProducersErr)
	} else {
		nConsumers, err := strconv.Atoi(os.Args[1])
		nTopics, err := strconv.Atoi(os.Args[2])

		if err != nil {
			log.Fatalln(ArgsErr)
		}

		for i := 0; i < nConsumers; i++ {
			go produce(nTopics)
		}

		c := make(chan os.Signal, 1)

		signal.Notify(c, os.Interrupt)

		<-c
	}
}
