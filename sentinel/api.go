package main

import (
	"fmt"
	"log"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"strconv"
	"bytes"
	"io/ioutil"
)

const (
	CREATE_NODE_URL = "/v1/catalog/register"
	GET_NODES_URL = "/v1/catalog/nodes"
	DELETE_NODES_URL = "/v1/catalog/deregister"
	CONTENT_TYPE = "application/json"
)

const nTopics int = 3
const nReplicas int = 2
var sentinel Sentinel

type TopicMeta struct {
	Title          string
	StorageAddress [nReplicas]string
	StorageNumber  int
	LastMessageAt  int
	MessagesHash   string
}

type StorageMeta struct {
	TopicsNumber int
	Address      string
	Topics       [nTopics]string
	Status       bool
	Id           string
}

type Node struct {
	Node string
	Address string
	Meta map[string]string
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


func createStorage(w http.ResponseWriter, r *http.Request) {
	var storage StorageMeta
	_ = json.NewDecoder(r.Body).Decode(&storage)

	fmt.Println(storage)
	
	var node Node
	node.Address = storage.Address
	node.Node = storage.Id
	node.Meta = make(map[string]string)
	for i := 0; i < storage.TopicsNumber; i++ {
		node.Meta["topics" + strconv.Itoa(i)] = storage.Topics[i]
	}

	fmt.Println(node)

	statusCode := saveConsulNode(node)

	w.WriteHeader(statusCode)
}

func getStorage(w http.ResponseWriter, r *http.Request) {
	field := r.FormValue("field")
	value := r.FormValue("value")
	getConsulNode(field, value, w)
}

func deleteStorage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	url := "http://localhost:8500" + DELETE_NODES_URL

	deleteMap := make(map[string]string)
	deleteMap["Node"] = id
	requestBody, err := json.Marshal(deleteMap)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(requestBody))
	if err != nil {
		// handle error
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}

	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)

}

func saveConsulNode(node Node) int {
	requestBody, err := json.Marshal(node)

	if err != nil {
		log.Fatal(err)
	}

	url := "http://localhost:8500" + CREATE_NODE_URL

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(requestBody))
	if err != nil {
		// handle error
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}

	defer resp.Body.Close()

	return resp.StatusCode
}

func getConsulNode(field string, value string, w http.ResponseWriter) {
	url := "http://localhost:8500" + GET_NODES_URL + "?filter=" + field + "==" + value

	resp, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	w.Write(body)
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


func main() {
	sentinel = Sentinel{}
	sentinel.Topics = make(map[string]TopicMeta)
	sentinel.Storages = make(map[string]StorageMeta)
	router := mux.NewRouter().StrictSlash(true)
	// router.HandleFunc("/rebalance", store)
	// go checkForBalance()
	router.HandleFunc("/storages", createStorage).Methods("POST")
	router.HandleFunc("/storages", getStorage).Methods("GET")
	router.HandleFunc("/storages/{id}", deleteStorage).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
}

