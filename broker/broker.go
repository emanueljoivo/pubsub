package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

const (
	ReplicatorAddress = "localhost:5000"
	PublisherPort     = "8000"
	SubscriberPort    = "9000"
	NetworkType       = "tcp"
	ContentType       = "application/json"
)

var (
	RequestError  = errors.New("request could not be completed")
	ConnectionErr = errors.New("connection error")
	UnmarshalErr  = errors.New("unmarshal error")
)

type TopicMessage struct {
	Topic     string
	Message   string
	CreatedAt time.Time
}

const MaxTopicsAmount int = 3

var topics map[string]string
var storagesAddr map[string]int8

type SubMessage struct {
	Topic  string
	Offset int
}

func dispatchMessage(cMsg <-chan TopicMessage) {
	for {
		message := <-cMsg

		topic := message.Topic

		for {
			getStorageUrl := ReplicatorAddress + "/storages/leader?topicName=" + topic

			resp, err := http.Get(getStorageUrl)

			storageData := make(map[string]string)
			err = json.NewDecoder(resp.Body).Decode(&storageData)

			if err != nil {
				log.Println(RequestError.Error())
			}

			storeEndpoint := storageData["Address"] + "/store"
			requestBody, err := json.Marshal(message)
			_, err = http.Post(storeEndpoint, ContentType, bytes.NewBuffer(requestBody))

			if err != nil {
				break
			}
		}
	}
}

func handleSub(c net.Conn, cMsg chan<- SubMessage) {
	for {
		message, err := bufio.NewReader(c).ReadBytes('\n')
		fmt.Println(message)

		if err != nil {
			log.Fatalln(ConnectionErr.Error() + " from sub")
		}

		var msg SubMessage
		err = json.Unmarshal(message, &msg)

		if err != nil {
			log.Println(UnmarshalErr.Error() + "from sub")
		}

		topic := msg.Topic
		getStorageUrl := ReplicatorAddress + "/storage?topicName=" + topic

		if err != nil {
			log.Fatal(err)
		}

		resp, err := http.Get(getStorageUrl)

		storageData := make(map[string]string)
		json.NewDecoder(resp.Body).Decode(&storageData)

		getTopicMessagesEndpoint := storageData["Address"] + "/get/" + topic + "?offset=" + strconv.Itoa(msg.Offset)
		response, err := http.Get(getTopicMessagesEndpoint)

		topicMessages := make(map[string]string)
		json.NewDecoder(response.Body).Decode(&storageData)

		e, err := json.Marshal(topicMessages["Messages"])

		if err != nil {
			log.Println(ConnectionErr)
		}
		_, err = c.Write(append(e, '\n'))
	}
}

func handlePub(c net.Conn, cMsg chan TopicMessage) {
	for {
		message, err := bufio.NewReader(c).ReadBytes('\n')

		if err != nil {
			log.Fatalln(ConnectionErr.Error() + " from pub")
		}

		var msg TopicMessage
		err = json.Unmarshal(message, &msg)

		msg.CreatedAt = time.Now().Local()

		if err != nil {
			log.Println(UnmarshalErr.Error() + "from pub")
		}
		// maybe select across various channels based on the topic?
		log.Printf("Msg %s enqueued\n", msg.Message)
		cMsg <- msg
	}
}

func publish(cMsg chan TopicMessage) {
	log.Println("Initializing pub broker")
	l, err := net.Listen(NetworkType, ":"+PublisherPort)
	if err != nil {
		log.Println(err)
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatalln(err)
			return
		}
		go handlePub(c, cMsg)
	}
}

func subscriber(cMsg chan SubMessage) {
	log.Println("Initializing sub broker")
	l, err := net.Listen(NetworkType, ":"+SubscriberPort)
	if err != nil {
		log.Println(err)
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatalln(err)
			return
		}
		go handleSub(c, cMsg)
	}
}


func main() {
	// the channel should be buffered
	pubMessages := make(chan TopicMessage, 100)

	// pub
	go publish(pubMessages)

	subMessages := make(chan SubMessage, 100)
	// sub
	go subscriber(subMessages)

	go dispatchMessage(pubMessages)

	cJoin := make(chan os.Signal, 1)

	signal.Notify(cJoin, os.Interrupt)

	<-cJoin
}
