package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"time"

	"github.com/gorilla/mux"
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
		panic(err)
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(messageJson)
	
}

func sub(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hey, this is sub")
	request := TopicRequest{}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		panic(err)
	}

	requestJson, err := json.Marshal(request)
	if err != nil{
		panic(err)
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(requestJson)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/pub", pub)
	router.HandleFunc("/sub", sub)

	log.Fatal(http.ListenAndServe(":8080", router))
}

