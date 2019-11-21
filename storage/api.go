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
	// "../structs"
	"strconv"
	"github.com/fatih/structs"
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

//VARS
const(
	nTopics int = 3
	nMessages int = 5
	timeToLive int = 10 //Seconds
)

var(
	port string
	sentinelPort string
	storage Storage
)

//FUNCS
func setupVariables() {
	port = os.Getenv("PORT")
	if port == "" {
	port = "8003"
}
	sentinelPort = os.Getenv("SENTINEL")
	fmt.Println(sentinelPort)
	if sentinelPort == "" { //This dont work, i dont know why
		sentinelPort = "8002"
	}
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
	return index,topic
}

func storeMessage(topicMessage TopicMessage) (int,TopicMeta) {
	topicName := topicMessage.Topic
	index,topic := getTopic(topicName)
	
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
	retries, _ := strconv.Atoi(result["Retries"])
	leader, _ := strconv.ParseBool(result["Leader"])
	topicMessage := TopicMessage{result["Topic"], result["Message"], createdAt}
	ans, meta := storeMessage(topicMessage)
	if (ans == -1) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w,"sorry, we cant create more topics ")
		return
	}

	// // if !leader {
	// 	propagate(topicMessage)
	// 	//treat error
	// // }

	updateSentinel(meta)
	//treat error
	topicMetaJson, err := json.Marshal(meta)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type","application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(topicMetaJson)
}

func propagate(message TopicMessage, retries int, meta TopicMeta) int { //Return an error status
	adresses := getOtherStorages(message.Topic)
	data := structs.Map(message)
	data["Leader"] = "1"
	data["Retries"] = retries
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

func updateSentinel(meta TopicMeta) {
	metaJson, err := json.Marshal(meta)
	if err != nil {
		panic(err)
	}
	_,err = http.Post("http://sentinel" + sentinelPort + "/topicAdress/" + meta.Topic,"application/json", bytes.NewBuffer(metaJson))
	if err != nil {
		panic(err)
	}
}

func getOtherStorages(topicName string) [2]string {
	// r, err = http.GET("http://sentinel" + sentinelPort + "/topicAdress/" + meta.Topic)
	return [2]string{"a","b"}	
}


func pingSentinel() {
	fmt.Println(sentinelPort)
	_, err := http.Get("http://sentinel" + ":"+sentinelPort +"/newStorage/"+port)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("hey, just pinged sentinel!")
}

func main() {
	setupVariables()
	// pingSentinel()	
	storage = Storage{}
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/store", store).Methods("POST")
	fmt.Println("Listening on port " + port)
	log.Fatal(http.ListenAndServe(":" + port, router))
}

