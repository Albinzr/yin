package util

import (
	"flag"
	"os"
	"runtime"
	"strconv"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

//Config :- env struct
type Config struct {
	Port       string
	KafkaURL   string
	KafkaTopic string
	GroupID    string
	Partition  int
	MinBytes   int
	MaxBytes   int
	FileSize   int64
	Retry      int
}

//LogError :- common function for loging error
func LogError(message string, errorData error) {
	if errorData != nil {
		log.Errorln("Error : ", message)
		return
	}
}

//LogInfo :- common func for loging info
func LogInfo(args ...interface{}) {
	log.Info(args)
}

//LogFatal :- common func for fatal error
func LogFatal(args ...interface{}) {
	log.Fatal(args)
}

//LogDebug :- common debug logger
func LogDebug(args ...interface{}) {
	log.Debug(args)
}

//LoadEnvConfig :- for loading config files
func LoadEnvConfig() *Config {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var err error
	var partition int
	var minBytes int
	var maxBytes int
	var fileSize int64
	var retry int

	key := flag.String("env", "development", "")
	flag.Parse()
	LogInfo("env:", *key)
	if *key == "production" {
		log.SetFormatter(&log.TextFormatter{})
		err = godotenv.Load("./production.env")
	} else {
		err = godotenv.Load("./local.env")
		log.SetFormatter(&log.TextFormatter{})
	}

	if err != nil {
		LogFatal("cannot load config file", err)
	}

	partition, err = strconv.Atoi(os.Getenv("PARTITION"))
	if err != nil {
		partition = 0
		LogError("unable to read from env key:PARTITION", err)
	}

	minBytes, err = strconv.Atoi(os.Getenv("MIN_BYTES"))
	if err != nil {
		minBytes = 0
		LogError("unable to read from env key:MIN_BYTES", err)
	}

	maxBytes, err = strconv.Atoi(os.Getenv("MAX_BYTES"))
	if err != nil {
		maxBytes = 1000000
		LogError("unable to read from env key:MAX_BYTES", err)
	}

	fileSize, err = strconv.ParseInt(os.Getenv("FILE_SIZE"), 10, 64)
	if err != nil {
		fileSize = 1
		LogError("unable to read from env key:FILE_SIZE", err)
	}
	retry, err = strconv.Atoi(os.Getenv("RETRY"))
	if err != nil {
		retry = 3
		LogError("unable to read from env key:RETRY", err)
	}

	config := new(Config)
	config.Port = os.Getenv("PORT")
	config.KafkaURL = os.Getenv("KAFKA")
	config.KafkaURL = os.Getenv("KAFKA_URL")
	config.KafkaTopic = os.Getenv("KAFKA_TOPIC")
	config.GroupID = os.Getenv("GROUP_ID")
	config.Partition = partition
	config.MinBytes = minBytes
	config.MaxBytes = maxBytes
	config.FileSize = fileSize
	config.Retry = retry
	return config
}
