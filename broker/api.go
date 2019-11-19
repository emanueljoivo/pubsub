package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

const SentinelAddress string = "localhost:5002"

var (
	RequestError = errors.New("request could not be completed")
)

type TopicMessage struct {
	Topic string
	Message string
	CreatedAt time.Time
}

type TopicRequest struct {
	Topic string
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
	if err != nil{
		log.Println(RequestError)
	}

	w.Header().Set("Content-Type","application/json")
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
	if err != nil{
		log.Println(RequestError)
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(requestJson)
}

func main() {
	time.Sleep(5*time.Second)
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/pub", pub)
	router.HandleFunc("/sub", sub)

	log.Fatal(http.ListenAndServe(":8000", router))
}

