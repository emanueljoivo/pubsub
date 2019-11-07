package main

import (
	"fmt"
	"log"
	"net/http"
	// "time"
	"encoding/json"
	"github.com/gorilla/mux"
)

const nTopics int = 3
const nReplicas int = 2
var sentinel Sentinel

type TopicMeta struct {
	Title string
	StorageAdress [nReplicas]string
	StorageNumber int
	LastMessageAt int
	MessagesHash string
}

type StorageMeta struct {
	TopicsNumber int
	Adress string
	Topics [nTopics]string
	Status bool
}

type Sentinel struct {
	Topics map[string]TopicMeta
	Storages map[string]StorageMeta
}

type TopicMessage struct {
	Topic string
	Message string
	CreatedAt int
}


func newStorage(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	port := params["port"]
	fmt.Println("HEYY " + port)

	storage := StorageMeta{}
	storage.Adress = port
	storage.Status = true
	storage.TopicsNumber = 0
	sentinel.Storages[port] = storage
	w.WriteHeader(http.StatusCreated)
}

func get(w http.ResponseWriter, r *http.Request) {
	sentinelJson, err := json.Marshal(sentinel)
	if err != nil{
		panic(err)
	}
	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(sentinelJson)
}

// func checkForBalance() {
// 	fmt.Println("just checking if everybody is ok")
// 	for ;; {
// 		fmt.Println("checking123")
// 		time.Sleep(5*time.Second)
// 	}
// }

func main() {
	sentinel = Sentinel{}
	sentinel.Topics = make(map[string]TopicMeta)
	sentinel.Storages = make(map[string]StorageMeta)
	router := mux.NewRouter().StrictSlash(true)
	// router.HandleFunc("/rebalance", store)
	// go checkForBalance()
	router.HandleFunc("/newStorage/{port}", newStorage)
	router.HandleFunc("/get", get)
	log.Fatal(http.ListenAndServe(":8002", router))
}

