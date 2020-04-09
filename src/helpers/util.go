package util

import (
	"flag"
	"os"
	"runtime"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Port string
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

	config := new(Config)
	config.Port = os.Getenv("PORT")
	return config
}
