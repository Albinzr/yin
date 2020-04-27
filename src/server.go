package server

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	util "applytics.in/yin/src/helpers"
	middleware "applytics.in/yin/src/middlewares"

	kafka "github.com/Albinzr/kafkaGo"
	queue "github.com/Albinzr/queueGo"

	engineio "github.com/googollee/go-engine.io"
	"github.com/googollee/go-engine.io/transport"
	"github.com/googollee/go-engine.io/transport/websocket"
	socket "github.com/googollee/go-socket.io"
)

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
		util.LogInfo("connected....:", s.RemoteAddr().String())
		s.Emit("ack", s.RemoteAddr().String())
		return nil
	})
}

func socketCloseListener(io *socket.Server) {
	io.OnDisconnect("/", func(s socket.Conn, msg string) {
		util.LogInfo("closed....:", s.RemoteAddr().String(), msg)
		s.Close()
	})
}

func socketBeaconListener(callback Message) {
	io.OnEvent("/", "beacon", func(s socket.Conn, msg string) {
		s.Emit("msgAck", "Recived msg")
		// util.LogInfo(msg)
		callback(msg + "\n")
	})
}

func beaconWriterCallback(message string) {
	fmt.Print(".")
	queueConfig.Insert(message)
}

func readQueueCallback(message string, fileName string) {
	util.LogInfo(fileName)

	kafkaConfig.WriteBulk(message, func(isWritten bool) {
		if isWritten {
			util.LogInfo("commiting")
			queueConfig.CommitFile(fileName)
			util.LogInfo("commited file: ", fileName)
			return
		} else {
			util.LogError("Cannot write to kafka: "+fileName, errors.New(""))
		}
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
