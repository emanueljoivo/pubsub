package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

const (
	ConsulAddr = "http://localhost:8500"
	RegisterServiceUri   = "/v1/agent/service/register"
	DeregisterServiceUri = "/v1/agent/service/deregister/:service_id"
	GetServiceHealthUri  = "/agent/health/service/id/:service_id"
	ContentType          = "application/json"
)

var (
	UnmarshalErr  = errors.New("unmarshal error")
)

const nTopics int = 3
const nReplicas int = 2

var Topics map[string]TopicMeta
var Storages map[string]StorageMeta

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
	Node    string
	Address string
	Meta    map[string]string
}

type TopicMessage struct {
	Topic     string
	Message   string
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
		node.Meta["topics"+strconv.Itoa(i)] = storage.Topics[i]
	}

	fmt.Println(node)

	statusCode := RegisterStorage(node)

	w.WriteHeader(statusCode)
}

func getStorage(w http.ResponseWriter, r *http.Request) {
	// get a available storage service meta in consul
	// returns to the requester
}

func DeregisterStorage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	url := ConsulAddr + DeregisterServiceUri

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

func RegisterStorage(node Node) int {
	requestBody, err := json.Marshal(node)

	if err != nil {
		log.Fatal(err)
	}

	url := ConsulAddr + RegisterServiceUri

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

func RegisterService(w http.ResponseWriter, r *http.Request) {
	var s StorageMeta
	body, _ := ioutil.ReadAll(r.Body)

	err := json.Unmarshal(body, &s)

	if err != nil {
		log.Printf(UnmarshalErr.Error())
	}
}

func getLeader(w http.ResponseWriter, r *http.Request) {

}

func getVersion(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(rw).Encode(`{"version":"0.0.1"}`); err != nil {
		log.Println(err.Error())
	}
}

func main() {
	Topics = make(map[string]TopicMeta)
	Storages = make(map[string]StorageMeta)

	router := mux.NewRouter()

	router.HandleFunc("/storages/register", RegisterService).Methods(http.MethodPost)
	router.HandleFunc("/storages/leader", getLeader).Methods(http.MethodGet)
	router.HandleFunc("/storages", createStorage).Methods(http.MethodPost)
	router.HandleFunc("/storages", getStorage).Methods(http.MethodGet)
	router.HandleFunc("/storages/{id}", DeregisterStorage).Methods(http.MethodDelete)
	router.HandleFunc("/version", getVersion).Methods(http.MethodGet)
	log.Fatal(http.ListenAndServe(":8080", router))
}
