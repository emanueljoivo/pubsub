package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
	"fmt"
)

const (
	ConsulAddr           = "http://127.0.0.1:8500"
	RegisterServiceUri   = "/v1/agent/service/register"
	DeregisterServiceUri = "/v1/agent/service/deregister/"
	GetServices          = "/v1/agent/services"
	GetService			 = "/agent/service/"
	GetServiceHealthUri  = "/agent/health/service/id/"
	ContentType          = "application/json"
)

var (
	UnmarshalErr  = errors.New("unmarshal error")
	RequestError = errors.New("request could not be completed")
)

var Topics map[string]TopicMeta
var Storages map[string]StorageMeta
var storagesCounter = 0

const nTopics int = 3
const nReplicas int = 2

type TopicMeta struct {
	Title          string
	StorageAddress [nReplicas]string
	StorageNumber  int
	LastMessageAt  int
	MessagesHash   string
}

type StorageMeta struct {
	Name 		 string
	TopicsNumber int
	Address      string
	Port		 int
	Topics       [nTopics]string
	Status       bool
	ID           string
}

type Check struct {
	HTTP string
	Interval string
	TTL string
}

type TopicMessage struct {
	Topic     string
	Message   string
	CreatedAt int
}

func GetStorage(w http.ResponseWriter, r *http.Request) {
	topicName := r.FormValue("topicName")
	log.Printf("Querying %s", topicName)

	url := ConsulAddr + GetService + topicName

	var sm StorageMeta

	response, err := http.Get(url)
	log.Println("responde " + response.Status)
	if err != nil {
		log.Println(RequestError.Error())
	}

	if response != nil {
		_ = json.NewDecoder(response.Body).Decode(&sm)
	}

	log.Println(sm)
	storageMeta, _ := json.Marshal(response)
	b, err := w.Write(storageMeta)

	if b == 0 {
		log.Println(RequestError.Error())
	}

}

func GetStorages(w http.ResponseWriter, r *http.Request) {
	url := ConsulAddr + GetServices
	var sms map[string]StorageMeta

	response, err := http.Get(url)
	log.Println(response)
	if err != nil {
		log.Println(RequestError.Error())
	}

	if response != nil {
		_ = json.NewDecoder(response.Body).Decode(&sms)
	}

	log.Println(sms)
	storesList, _ := json.Marshal(response)
	b, err := w.Write(storesList)

	if b == 0 {
		log.Println(RequestError.Error())
	}
}

func RegisterService(w http.ResponseWriter, r *http.Request) {
	var s map[string]string
	_ = json.NewDecoder(r.Body).Decode(&s)
	fmt.Println(s)
	url := ConsulAddr + RegisterServiceUri

	var ss StorageMeta
	var absoluteAddr = s["Address"]
	l := strings.Split(absoluteAddr, ":")

	ss.Address = l[0]
	ss.Port, _ = strconv.Atoi(l[1])
	ss.Name = "storage_" + strconv.Itoa(storagesCounter)
	ss.ID =  s["ID"]

	storagesCounter++

	Storages[ss.Name] = ss

	log.Print(ss)
	bd, _ := json.Marshal(ss)

	c := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(bd))

	if err != nil {
		log.Println(RequestError.Error())
	}

	resp, err := c.Do(req)
	log.Println(resp.Status)
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
}

func Leader(rw http.ResponseWriter, req *http.Request) {
	topicName := req.FormValue("topicName")
	log.Printf("Querying %s leader", topicName)

	// search/elect leader
	// retrieves leader
}

func Version(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(rw).Encode(`{"version":"0.0.1"}`); err != nil {
		log.Println(err.Error())
	}
}

func main() {
	Topics = make(map[string]TopicMeta)
	Storages = make(map[string]StorageMeta)
	log.Print("Sentinel starting\n")
	router := mux.NewRouter()

	router.HandleFunc("/storages/register", RegisterService).Methods(http.MethodPost)

	router.HandleFunc("/storages/leader", Leader).Methods(http.MethodGet)
	router.HandleFunc("/storage", GetStorage).Methods(http.MethodGet)
	router.HandleFunc("/storages", GetStorages).Methods(http.MethodGet)

	router.HandleFunc("/storages/{id}", DeregisterStorage).Methods(http.MethodDelete)

	router.HandleFunc("/version", Version).Methods(http.MethodGet)
	log.Fatal(http.ListenAndServe(":8080", router))
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
