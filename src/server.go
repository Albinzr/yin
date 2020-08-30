package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	cache "applytics.in/yin/src/cache"
	util "applytics.in/yin/src/helpers"

	kafka "github.com/Albinzr/kafkaGo"
	queue "github.com/Albinzr/queueGo"
	socket "github.com/Albinzr/socketGo"
)

type sessionConfig struct {
	ShouldRecord bool        `json:"shouldRecord"`
	Config       interface{} `json:"config"`
}

//Message :- simple type for message callback
type Message func(message string)

var env = util.LoadEnvConfig()
var path, _ = filepath.Abs("./store")

var socketConfig = &socket.Config{
	Network:      "tcp",
	Address:      ":1000",
	OnConnect:    onConnect,
	OnDisconnect: onDisonnect,
	OnRecive:     onRecive,
}

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
	go func() {
		log.Fatal(http.ListenAndServe(":1001", nil))
	}()
	//log
	logStartDetails()

	//Start reading msgs from file and pass it to kafka
	go readMessageToKafka()

	//configs
	queueConfig.Init()
	cacheConfig.Init()
	socketConfig.Init()

}

func logStartDetails() {
	fmt.Printf("%+v\n", kafkaConfig)
	fmt.Printf("%+v\n", queueConfig)
	util.LogInfo("Temp file storage path: ", path)
	util.LogInfo("Env: ", env)
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

func onConnect(s *socket.Socket) {

	IP := s.IP
	util.LogInfo("connected....:", IP)
	aID := s.Aid
	cacheConfig.UpdateOnlineCount(aID)

	config := &sessionConfig{
		ShouldRecord: true,
		Config:       nil,
	}
	configJSON, err := json.Marshal(config)
	if err != nil {
		util.LogError("cannot create json from config", err)
	}
	fmt.Println("************** On Connection Start**************")
	fmt.Println(string(configJSON))
	fmt.Println("****************************")
	s.Write(string(configJSON))

}

func onDisonnect(s *socket.Socket) {
	util.LogInfo("closed....:", s.IP)

	cacheConfig.ReduceOnlineCount(s.Aid)

	closeJSON, err := json.Marshal(s)
	fmt.Println("**************On Conccection End **************")
	fmt.Println(string(closeJSON))
	fmt.Println("****************************")

	if err != nil {
		util.LogError("could not create close json", err)
	}

	closeMsg := string(closeJSON) + "\n"
	beaconWriterCallback("dc " + closeMsg)
	PrintMemUsage()
}

func onRecive(s *socket.Socket, channel string, msg string) {
	if channel == "/beacon" {
		beaconWriterCallback("en " + msg + "\n")
	} else {
		beaconWriterCallback("dc " + msg + "\n")
	}

}

func beaconWriterCallback(message string) {
	queueConfig.Insert(message)
}

func readQueueCallback(message string, fileName string) {
	util.LogInfo("reading files")

	kafkaConfig.WriteBulk(message, func(isWritten bool) {
		if isWritten {
			queueConfig.CommitFile(fileName)
			util.LogInfo("committed file: ", fileName)
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
	err := errors.New("cannot connect to kafka")
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
	fmt.Printf("\tMemory Freed = %v\n", bToMb(m.Frees))

	runtime.GC()
	debug.FreeOSMemory()
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
