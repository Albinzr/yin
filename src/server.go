package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	cache "applytics.in/yin/src/cache"
	util "applytics.in/yin/src/helpers"
	middleware "applytics.in/yin/src/middlewares"

	kafka "github.com/Albinzr/kafkaGo"
	queue "github.com/Albinzr/queueGo"

	engineio "github.com/googollee/go-engine.io"
	"github.com/googollee/go-engine.io/transport"
	"github.com/googollee/go-engine.io/transport/websocket"
	socket "github.com/googollee/go-socket.io"
)

type closeMessage struct {
	endTime time.Duration
	ip      string
	aid     string
	sid     string
	status  string
}

//Message :- simple type for message callback
type Message func(message string)

var env = util.LoadEnvConfig()
var path, _ = filepath.Abs("./store")
var io *socket.Server = setupSocket()
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

	//Socket io connection event listener
	socketConnectionListener()

	//Socket io beacon listner
	socketBeaconListener(beaconWriterCallback)

	//Socket io connection close listener
	socketCloseListener(io)

	setupHTTPServer(env.Port, io)
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

func setupSocket() *socket.Server {

	transporter := websocket.Default
	transporter.CheckOrigin = func(req *http.Request) bool {
		return true
	}

	options := &engineio.Options{Transports: []transport.Transport{transporter}}
	server, err := socket.NewServer(options)
	util.LogError("cannot start socket server", err)

	go server.Serve()
	return server
}

func setupHTTPServer(port string, io *socket.Server) {
	http.Handle("/socket.io/", middleware.EnableCors(io))
	util.LogInfo("Serving at localhost:" + port)
	util.LogFatal(http.ListenAndServe(":"+port, nil))
	defer io.Close()
}

func socketConnectionListener() {
	io.OnConnect("/", func(s socket.Conn) error {

		IP := s.RemoteHeader().Get("X-Real-Ip")
		util.LogInfo("connected....:", IP)

		query := s.URL().RawQuery
		querySplit := strings.Split(query, "&")
		sidQuery := querySplit[0]
		aidQuery := querySplit[1]
		sID := strings.Split(sidQuery, "=")[1]
		aID := strings.Split(aidQuery, "=")[1]

		cacheConfig.AddAppID(aID)
		cacheConfig.AddIP(IP, sID)
		s.Emit("ack", IP)
		return nil
	})
}

func socketCloseListener(io *socket.Server) {
	io.OnDisconnect("/", func(s socket.Conn, msg string) {
		IP := s.RemoteHeader().Get("X-Real-Ip")
		util.LogInfo("closed....:", IP)
		cacheConfig.RemoveIP(IP)

		query := s.URL().RawQuery
		querySplit := strings.Split(query, "&")
		sidQuery := querySplit[0]
		aidQuery := querySplit[1]
		sID := strings.Split(sidQuery, "=")[1]
		aID := strings.Split(aidQuery, "=")[1]

		cacheConfig.RemoveAppID(aID)

		close := &closeMessage{
			status:  "close",
			sid:     sID,
			aid:     aID,
			ip:      IP,
			endTime: time.Nanosecond,
		}

		closeJSON, err := json.Marshal(close)

		if err != nil {
			util.LogError("could not create close json", err)
		}

		msg := string(closeJSON) + "\n"

		beaconWriterCallback(msg)

		s.Close()
	})
}

func socketBeaconListener(callback Message) {
	io.OnEvent("/", "beacon", func(s socket.Conn, msg string) {
		s.Emit("msgAck", "Recived msg")
		callback(msg + "\n")
	})
}

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
