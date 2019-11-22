package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	// "time"
	"os"
	"time"
)

const nTopics int = 3
const nMessages int = 5
const TTL int = 10 //Seconds

const (
	ServerPortEnvK    = "SERVER_PORT"
	SentinelHostEnvK  = "SENTINEL_HOST"
	SentinelPortEnvK    = "SENTINEL_PORT"
	DefaultServerPort = "8082"
	ContentType          = "application/json"
)

var ServerPort string
var SentinelPort string
var SentinelHost string

func setupVariables() {
	if p, exists := os.LookupEnv(ServerPortEnvK); !exists {
		ServerPort = DefaultServerPort
	} else {
		ServerPort = p
	}
	log.Printf("Server post %s:",ServerPort)
	
	if h, exists := os.LookupEnv(SentinelHostEnvK); !exists {
		log.Fatalln("Require specify the sentinel host")
	} else {
		SentinelHost = h
	}	
	log.Printf("Sentinel host %s", SentinelHost)

	if p, exists := os.LookupEnv(SentinelPortEnvK); !exists {
		log.Fatalln("Require specify the sentinel http port")
	} else {
		SentinelPort = p
	}
	log.Printf("Sentinel port %s:",ServerPort)
	
}
func computeHashKeyForList(list [5]string) string {
	var buffer bytes.Buffer
	for i := range list {
		buffer.WriteString(list[i])
	}
	hash := sha256.Sum256([]byte(buffer.String()))
	hashString := string(hash[:])
	return hashString
}

func getMeta(topic Topic) TopicMeta {
	meta := TopicMeta{}
	meta.Hash = topic.Hash
	meta.Title = topic.Title
	meta.LastMessageAt = topic.LastMessageAt
	return meta
}

type TopicMessage struct {
	Topic     string
	Message   string
	CreatedAt int
}

type Topic struct {
	Messages      [nMessages]string //Last message is the newest
	Title         string
	LastMessageAt int
	Hash          string
}

type TopicMeta struct {
	Title         string
	Hash          string
	LastMessageAt int
}

type Storage struct {
	Topics  [nTopics]Topic
	nTopics int
}

var storage Storage

func getTopic(topicName string) (int, Topic) {
	index := -1
	var topic Topic
	for i, storageTopic := range storage.Topics {
		if storageTopic.Title == topicName {
			topic = storageTopic
			fmt.Println("HEY U ALREADY HAVE ONE")
			index = i
		}
	}
	if index == -1 {
		topic.Title = topicName
		index = storage.nTopics
		storage.nTopics++
	}
	if index >= nTopics {
		index = -1
	}
	return index, topic
}

func store(w http.ResponseWriter, r *http.Request) {
	topicMessage := TopicMessage{}

	err := json.NewDecoder(r.Body).Decode(&topicMessage)
	if err != nil {
		panic(err)
	}
	topicName := topicMessage.Topic
	index, topic := getTopic(topicName)

	if index == -1 {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "sorry, we cant create more topics ")
		return
	}

	// Shifting the message queue
	// There must be a better way of doing this. From here
	var slice []string
	slice = append(topic.Messages[1:5], topicMessage.Message)
	copy(topic.Messages[:], slice[0:5])
	// to here.

	topic.LastMessageAt = topicMessage.CreatedAt
	topicHash := computeHashKeyForList(topic.Messages)
	topic.Hash = topicHash
	storage.Topics[index] = topic

	topicMeta := getMeta(topic)
	topicMetaJson, err := json.Marshal(topicMeta)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(topicMetaJson)
}

func wakeup() {
	// url := "http://" + SentinelHost + ":" + SentinelPort + "/storages/register"
	//_, err := http.Post(url, ContentType,)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	log.Println("hey, just pinged sentinel!")
}

func init() {
	setupVariables()
	wakeup()
}

func main() {
	time.Sleep(7 * time.Second)

	storage = Storage{}
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/store", store)
	log.Fatal(http.ListenAndServe(":"+ServerPort, router))
}
