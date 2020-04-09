package server

import (
	"bufio"
	"errors"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	util "applytics.in/yin/src/helpers"
	kafka "applytics.in/yin/src/kafka"
	middleware "applytics.in/yin/src/middlewares"
	queue "applytics.in/yin/src/queue"

	engineio "github.com/googollee/go-engine.io"
	"github.com/googollee/go-engine.io/transport"
	"github.com/googollee/go-engine.io/transport/websocket"
	socket "github.com/googollee/go-socket.io"
)

//Message :- simple type for message callback
type Message func(message string)

var env = util.LoadEnvConfig()
var path, _ = filepath.Abs("../temp")
var io *socket.Server = setupSocket()

var kafkaConfig = &kafka.Config{
	Topic:     "beacon-1",
	Partition: 0,
	URL:       "localhost:9093",
	GroupID:   "beaconGroup-1",
	MinBytes:  1,
	MaxBytes:  10485760,
}

var queueConfig = &queue.Config{
	StoragePath: path,
	FileSize:    1,
	NoOfRetries: 3,
}

func Start() {

	//Start reading msgs from file and pass it to kafka
	go readMessageToKafka()

	//Socket io connection event listener
	socketConnectionListener()

	//Socket io beacon listner
	socketBeaconListener(beaconWriterCallback)

	//Socket io connection close listener
	socketCloseListener(io)

	setupHTTPServer(env.Port, io)
}

func readMessageToKafka() {
	util.LogInfo("Started reading message from file")

	//Start kafka
	err := startKafka()
	util.LogError("Kafka connection issue", err)

	if err != nil {
		//if error try after (T) sec
		time.AfterFunc(30*time.Second, readMessageToKafka)
		return
	}
	//Read msg from file
	readFromQueue()
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
		util.LogInfo("connected....:", s.RemoteAddr().String())
		s.Emit("ack", s.RemoteAddr().String())
		return nil
	})
}

func socketCloseListener(io *socket.Server) {
	io.OnDisconnect("/", func(s socket.Conn, msg string) {
		util.LogInfo("closed")
		s.Close()
	})
}

func socketBeaconListener(callback Message) {
	io.OnEvent("/", "beacon", func(s socket.Conn, msg string) {
		callback(msg)
	})
}

func beaconWriterCallback(message string) {
	path, _ := filepath.Abs("../temp")
	config := &queue.Config{
		StoragePath: path,
		FileSize:    1,
		NoOfRetries: 3,
	}
	config.Insert(message)
}

func readQueueCallback(message string, fileName string) {

	errCount := 0
	scanner := bufio.NewScanner(strings.NewReader(message))
	for scanner.Scan() {
		msg := scanner.Text()
		kafkaConfig.Write(msg, func(isWritten bool) {
			if !isWritten {
				errCount++
			}
		})
	}
	if errCount == 0 {
		queueConfig.CommitFile(fileName)
		return
	}
	util.LogError("Cannot write to kafka: "+fileName+"& count:"+strconv.Itoa(errCount), errors.New(""))
}

func readFromQueue() {
	//Read from file
	queueConfig.Read(readQueueCallback)
}

func startKafka() error {
	if kafkaConfig.Init() {
		util.LogInfo("Connected to kafka")
		return nil
	}
	err := errors.New("Cannot connect to kafka")
	return err
}
