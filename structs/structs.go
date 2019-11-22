package structs

const(
	nTopics int = 3
	nMessages int = 5
	timeToLive int = 10 //Seconds
	nReplicas int = 3
)


//Topic
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


//Broker
type TopicRequest struct {
	Topic  string
	Offset int
}

//Storage
type Storage struct {
	Topics [nTopics]Topic
	nTopics int
}

//Sentinel

type SentinelTopicMeta struct {
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
}

type Sentinel struct {
	Topics map[string]TopicMeta
	Storages map[string]StorageMeta
}