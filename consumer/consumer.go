package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"
)

const (
	SubPort     = "9000"
	NetworkType = "tcp"
)

var (
	ConnectionErr = errors.New("connection error")
	UnmarshalErr  = errors.New("unmarshal error")
	ConsumersErr = errors.New("you need to specify how many consumers will perform")
	ArgsErr = errors.New("args error")
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

func consume(nTopics int) {
	conn, err := net.Dial(NetworkType, ":"+SubPort)

	for err != nil {
		log.Println("Waiting connection with broker")
		time.Sleep(10 * time.Second)
	}

	defer conn.Close()

	log.Println("Initializing consumers on port :" + SubPort)

	for i := 0; i < nTopics; i++ {
		go handleConnection(conn, "topic_" + strconv.Itoa(i))
	}
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalln(ConsumersErr)
	} else {
		nConsumers, err := strconv.Atoi(os.Args[1])
		nTopics, err := strconv.Atoi(os.Args[2])

		if err != nil {
			log.Fatalln(ArgsErr)
		}

		for i := 0; i < nConsumers; i++ {
			go consume(nTopics)
		}

		c := make(chan os.Signal, 1)

		signal.Notify(c, os.Interrupt)

		<-c
	}
}
