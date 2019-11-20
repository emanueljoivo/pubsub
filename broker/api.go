package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

const (
	SentinelAddress = "localhost:5002"
	PublisherPort = "8000"
	SubscriberPort = "9000"
	NetworkType   = "tcp"
)

var (
	RequestError = errors.New("request could not be completed")
	ConnectionErr = errors.New("connection error")
	UnmarshalErr  = errors.New("unmarshal error")
)

type TopicMessage struct {
	Topic     string
	Message   string
	CreatedAt time.Time
}

type TopicRequest struct {
	Topic  string
	Offset int
}

func dispatchMsg(topic TopicMessage) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	_ = enc.Encode(topic)
	var jsonStr = buffer.Bytes()

	requestUrl := SentinelAddress + "/newStorage/" + strconv.Itoa(10)

	req, _ := http.NewRequest("POST", requestUrl, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Print(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

func pub(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hey, this is pub")
	message := TopicMessage{}

	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		panic(err)
	}

	message.CreatedAt = time.Now().Local()

	messageJson, err := json.Marshal(message)
	if err != nil {
		log.Println(RequestError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(messageJson)

}

func sub(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hey, this is sub")
	request := TopicRequest{}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println(err)
	}

	requestJson, err := json.Marshal(request)
	if err != nil {
		log.Println(RequestError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(requestJson)
}

func handleSub(c net.Conn, cMsg chan TopicMessage) {
	for {
		// maybe select across various channels based on the topic?
		message := <- cMsg

		e, err := json.Marshal(message)

		if err != nil {
			log.Println(ConnectionErr.Error() + "from sub")
		}

		_, err = c.Write(append(e, '\n'))
		log.Printf("Msg %s dequeued \n", message.Message)

		if err != nil {
			log.Fatalln(ConnectionErr.Error() + "from sub")
		}
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

		if err != nil {
			log.Println(UnmarshalErr.Error() + "from pub")
		}
		// maybe select across various channels based on the topic?
		log.Printf("Msg %s enqueued\n", msg.Message)
		cMsg <- msg
	}
}

func main() {
	// the channel should be buffered
	messages := make(chan TopicMessage, 100)

	// pub
	go func(cMsg chan TopicMessage) {
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
	} (messages)

	// sub
	go func(cMsg chan TopicMessage) {
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
	} (messages)

	cJoin := make(chan os.Signal, 1)

	signal.Notify(cJoin, os.Interrupt)

	<-cJoin
}
