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
	// "../structs"
	"strconv"
	"github.com/fatih/structs"
)


const (
	nTopics int = 3
	nMessages int = 5
	TTL int = 10 //Seconds
	ServerPortEnvK    = "SERVER_PORT"
	SentinelHostEnvK  = "SENTINEL_HOST"
	DefaultSentinelHost = "http://127.0.0.1"
	SentinelPortEnvK  = "SENTINEL_PORT"
	DefaultSentinelPort = "8080"
	DefaultServerPort = "8003"
	ContentType       = "application/json"
	ServerAdressEnvK  = "SERVER_ADDRESS"
	DefaultServerAdress = "localhost"
)

var(
	ServerAdress string
	ServerPort string
	SentinelPort string
	SentinelHost string
	port string
	sentinelPort string
	storage Storage
)

//STRUCTS
type TopicMessage struct {
	Topic string
	Message string
	CreatedAt int
}

type Topic struct {
	Messages [nMessages]string //Last message is the newest
	Topic string
	LastMessageAt int
	Hash string
}

type TopicMeta struct {
	Topic string
	Hash string
	LastMessageAt int
}

type Storage struct {
	Topics [nTopics]Topic
	nTopics int
}


//FUNCS
func setupVariables() {
	if p, exists := os.LookupEnv(ServerPortEnvK); !exists {
		ServerPort = DefaultServerPort
	} else {
		ServerPort = p
	}
	log.Printf("Server post %s:",ServerPort)
	
	if h, exists := os.LookupEnv(SentinelHostEnvK); !exists {
		SentinelHost = DefaultSentinelPort
	} else {
		SentinelHost = h
	}	
	log.Printf("Sentinel host %s", SentinelHost)

	if p, exists := os.LookupEnv(SentinelPortEnvK); !exists {
		SentinelPort = DefaultSentinelPort
	} else {
		SentinelPort = p
	}
	log.Printf("Sentinel port %s:",SentinelPort) //this could be fixed

	if p, exists := os.LookupEnv(DefaultServerAdress); !exists {
		ServerAdress = DefaultServerAdress
	} else {
		ServerAdress = p
	}
	log.Printf("Server adress %s:",ServerAdress)
	
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
	meta.Topic = topic.Topic
	meta.LastMessageAt = topic.LastMessageAt
	return meta
}

func getTopic(topicName string) (int, Topic) {
	index := -1
	var topic Topic
	topic.LastMessageAt = -1
	for i , storageTopic := range storage.Topics {
		if storageTopic.Topic == topicName {
			topic = storageTopic
			index = i
		}
	}
	if index == -1 {
		topic.Topic = topicName
		index = storage.nTopics
		storage.nTopics++
	}
	if index >= nTopics {
		index = -1
	}
	return index, topic
}

func storeMessage(topicMessage TopicMessage) (int,TopicMeta) {
	topicName := topicMessage.Topic
	index, topic := getTopic(topicName)

	if index == -1 {
		return -1, TopicMeta{}
	}
	if topic.LastMessageAt != topicMessage.CreatedAt {
		// Shifting the message queue
		// There must be a better way of doing this. From here
		var slice []string
		slice = append(topic.Messages[1:5],topicMessage.Message)
		copy(topic.Messages[:], slice[0:5])
		// to here.

		topic.LastMessageAt = topicMessage.CreatedAt
		topicHash := computeHashKeyForList(topic.Messages)
		topic.Hash = topicHash
		storage.Topics[index] = topic
	}
	return 1, getMeta(topic)
}
func store(w http.ResponseWriter, r *http.Request) {

	var result map[string]string

	_ = json.NewDecoder(r.Body).Decode(&result)
	createdAt, _ := strconv.Atoi(result["CreatedAt"])
	Leader, _ := strconv.ParseBool(result["Leader"])
	topicMessage := TopicMessage{result["Topic"], result["Message"], createdAt}
	ans, meta := storeMessage(topicMessage)

	if (ans == -1) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "sorry, we cant create more topics ")
		return
	}

	// // if !leader {
	// 	propagate(topicMessage)
	// 	//treat error
	// // }

	// updateSentinel(meta)
	//treat error

	topicMetaJson, err := json.Marshal(meta)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(topicMetaJson)
}

func wakeup() {
	adress := map[string]string{"Address": ServerAdress,
								"Port": ServerPort}
	adressJson, err := json.Marshal(adress)
	if err != nil {
		panic(err)
	}
	url := "http://" + SentinelHost + ":" + SentinelPort + "/storages/register"
	_, err = http.Post(url, ContentType, bytes.NewBuffer(adressJson))
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("hey, just pinged sentinel!")
}

func propagate(message TopicMessage, retries int, meta TopicMeta) int { //Return an error status
	adresses := getOtherStorages(message.Topic)
	data := structs.Map(message)
	data["Leader"] = "1"
	messageJson, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	for _, adress := range adresses {
		r , err := http.Post("http://" + adress + "/store/" + message.Topic,"application/json", bytes.NewBuffer(messageJson))
		if err != nil {
			log.Fatalln(err) //Treat error
		}

		var result map[string]string
		_ = json.NewDecoder(r.Body).Decode(&result)
		lastMessageAt, _ := strconv.Atoi(result["LastMessageAt"])
		hash := result["Hash"]
		if (lastMessageAt != meta.LastMessageAt && hash != meta.Hash) {
			//ERROR
			return -1
		}
	}
	return 1
}


func getOtherStorages(topicName string) [2]string {
	// r, err = http.GET("http://sentinel" + sentinelPort + "/topicAdress/" + meta.Topic)
	return [2]string{"a","b"}	
}


func getTopicLastMessage(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	topicName := params["topic"]
	_,topic := getTopic(topicName)
	lastMessage := topic.Messages[nMessages - 1]
	fmt.Fprintf(w,lastMessage)
}

func getAll(w http.ResponseWriter, r *http.Request) {
	storageJson, err := json.Marshal(storage)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(storageJson)

}

func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,"OK")
}

func init() {
	setupVariables()
	wakeup()
}

func main() {
	storage = Storage{}
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/store", store)
	router.HandleFunc("/get/{topic}", getTopicLastMessage).Methods("GET")
	router.HandleFunc("/get", getAll).Methods("GET")
	fmt.Println("Listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+ServerPort, router))
}
