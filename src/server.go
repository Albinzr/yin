package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	cache "applytics.in/yin/src/cache"
	util "applytics.in/yin/src/helpers"

	kafka "github.com/Albinzr/kafkaGo"
	queue "github.com/Albinzr/queueGo"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 5,
	WriteBufferSize: 1024 * 5,
}

//CloseMessage :- Close Message struct for end session
type CloseMessage struct {
	EndTime int64  `json:"endTime"`
	IP      string `json:"ip"`
	Aid     string `json:"aid"`
	Sid     string `json:"sid"`
	Status  string `json:"type"`
}

//Message :- simple type for message callback
type Message func(message string)

var env = util.LoadEnvConfig()
var path, _ = filepath.Abs("./store")

var kafkaConfig = &kafka.Config{
	Topic:     env.KafkaTopic,
	Partition: env.Partition,
	URL:       env.KafkaURL,
	GroupID:   env.GroupID,
	MinBytes:  env.MinBytes,
	MaxBytes:  env.MaxBytes,
}

var cacheConfig = &cache.Config{
	Host: "redis",
	Port: "6379",
	// Password: "",
}

var queueConfig = &queue.Config{
	StoragePath: path,
	FileSize:    env.FileSize,
	NoOfRetries: env.Retry,
}

//Start :- server start function
func Start() {
	fmt.Printf("%+v\n", kafkaConfig)
	fmt.Printf("%+v\n", queueConfig)

	util.LogInfo("Temp file storage path: ", path)
	util.LogInfo("Env: ", env)
	queueConfig.Init() // seprate thread
	cacheConfig.Init()

	//Start reading msgs from file and pass it to kafka
	go readMessageToKafka() // seprate thread

	//Socket io connection close listener
	// socketCloseListener(io)

	setupHTTPServer(env.Port)
}

func readMessageToKafka() {

	//Start kafka
	err := startKafka()
	if err != nil {
		util.LogError("Kafka connection issue", err)
		//if error try after (T) sec
		time.AfterFunc(30*time.Second, readMessageToKafka)
	}

	util.LogInfo("Started reading message from file")
	//Read from file
	queueConfig.Read(readQueueCallback)
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func setupHTTPServer(port string) {
	http.HandleFunc("/beacon", echo)
	util.LogFatal(http.ListenAndServe(":"+port, nil))
}

// func socketConnectionListener() {
// 	io.OnConnect("/", func(s socket.Conn) error {

// 		IP := s.RemoteHeader().Get("X-Real-Ip")
// 		util.LogInfo("connected....:", IP)

// 		query := s.URL().RawQuery
// 		querySplit := strings.Split(query, "&")
// 		aidQuery := querySplit[1]
// 		aID := strings.Split(aidQuery, "=")[1]

// 		cacheConfig.UpdateOnlineCount(aID)

// 		s.Emit("status", "connected")
// 		return nil
// 	})
// }

// func socketCloseListener(io *socket.Server) {
// 	io.OnDisconnect("/", func(s socket.Conn, msg string) {
// 		IP := s.RemoteHeader().Get("X-Real-Ip")
// 		util.LogInfo("closed....:", IP)

// 		query := s.URL().RawQuery
// 		querySplit := strings.Split(query, "&")
// 		sidQuery := querySplit[0]
// 		aidQuery := querySplit[1]
// 		sID := strings.Split(sidQuery, "=")[1]
// 		aID := strings.Split(aidQuery, "=")[1]

// 		cacheConfig.ReduceOnlineCount(aID)

// 		close := &CloseMessage{
// 			Status:  "close",
// 			Sid:     sID,
// 			Aid:     aID,
// 			IP:      IP,
// 			EndTime: time.Now().UnixNano() / int64(time.Millisecond),
// 		}

// 		util.LogInfo(sID, aID, IP, time.Nanosecond)
// 		closeJSON, err := json.Marshal(close)

// 		if err != nil {
// 			util.LogError("could not create close json", err)
// 		}

// 		closeMsg := string(closeJSON) + "\n"
// 		beaconWriterCallback(closeMsg)
// 		PrintMemUsage()
// 		closeErr := s.Close().Error()
// 		util.LogInfo(closeErr)
// 	})
// }

// func socketBeaconListener(callback Message) {
// 	io.OnEvent("/", "beacon", func(s socket.Conn, msg string) {
// 		ID := msg[0:5]
// 		util.LogInfo(ID)
// 		s.Emit("ack", ID)
// 		callback(msg[5:] + "\n")
// 	})
// }

// func socketBeaconEndListener(callback Message) {
// 	io.OnError("/", func(s socket.Conn, err error) {
// 		util.LogError("socket error", err)
// 	})

// 	io.OnEvent("/", "beaconEnd", func(s socket.Conn, msg string) {
// 		util.LogInfo(msg)
// 		callback(msg + "\n")
// 	})
// }

func beaconWriterCallback(message string) {
	fmt.Print(".")
	queueConfig.Insert(message)
}

func readQueueCallback(message string, fileName string) {
	util.LogInfo("reading files")

	kafkaConfig.WriteBulk(message, func(isWritten bool) {
		if isWritten {
			queueConfig.CommitFile(fileName)
			util.LogInfo("commited file: ", fileName)
			return
		}
		util.LogError("Cannot write to kafka: "+fileName, errors.New(""))
	})

}

func startKafka() error {
	if kafkaConfig.IsKafkaReady() {
		util.LogInfo("Connected to kafka")
		return nil
	}
	err := errors.New("Cannot connect to kafka")
	return err
}

//PrintMemUsage -test
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
