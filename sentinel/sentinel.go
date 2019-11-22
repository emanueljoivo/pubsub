package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	ConsulAddr = "http://127.0.0.1:8500"
	RegisterServiceUri   = "/v1/agent/service/register"
	DeregisterServiceUri = "/v1/agent/service/deregister/:service_id"
	GetServiceHealthUri  = "/agent/health/service/id/:service_id"
	ContentType          = "application/json"
)

var (
	UnmarshalErr  = errors.New("unmarshal error")
	RequestError = errors.New("request could not be completed")
)

var Topics map[string]TopicMeta
var Storages map[string]StorageMeta

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
	TopicsNumber int
	Address      string
	Port		string
	Topics       [nTopics]string
	Status       bool
	ID           string
}

type Check struct {
	HTTP string
	Interval string
	TTL string
}

type StorageService struct {
	ID string
	Name string
	Address string
	Port string
}

type TopicMessage struct {
	Topic     string
	Message   string
	CreatedAt int
}

func GetStorage(w http.ResponseWriter, r *http.Request) {
	// get a available storage service meta in consul
	// returns to the requester
}

func GetStorages(w http.ResponseWriter, r *http.Request) {

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

func RegisterService(w http.ResponseWriter, r *http.Request) {
	var s StorageMeta
	log.Println(s)
	body, _ := ioutil.ReadAll(r.Body)

	err := json.Unmarshal(body, &s)

	if err != nil {
		log.Printf(UnmarshalErr.Error())
	}

	url := ConsulAddr + RegisterServiceUri

	var ss StorageService
	ss.Name = "storage"
	ss.Port = s.Port
	ss.Address = s.Address
	ss.ID = s.ID

	log.Println(ss)

	reqBody, err := json.Marshal(&ss)

	c := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(reqBody))

	if err != nil {
		log.Println(RequestError.Error())
	}

	resp, err := c.Do(req)
	log.Print(resp)
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
