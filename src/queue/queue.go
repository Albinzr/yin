package queue

import (
	"io/ioutil"
	"os"
	"sync"
	"time"
)

var wg sync.WaitGroup

//Element :- struct formater
type Element struct {
	data    string
	retries int
}

//Config :- struct formater
type Config struct {
	queue      []*Element
	isWritting bool
	isReading  bool
	//user
	StoragePath string
	FileSize    int64
	NoOfRetries int
}

//Init :- main function of module
func (c *Config) Init() {
	c.queue = []*Element{}
	c.isWritting = false
	c.isReading = false
	c.FileSize = 1
	// c.noOfRetries = c.noOfRetries || 0
	// c.fileSize = c.fileSize || 0
	// c.storagePath = c.storagePath || "/temp"
}

//Insert : - export function of module
func (c *Config) Insert(data string) {
	ele := new(Element)
	ele.data = data
	ele.retries = 0

	c.queue = append(c.queue, ele)
	c.execute()
}

func (c *Config) execute() {
	if c.isWritting {
		return
	}
	c.isWritting = true
	c.processQueue()
}

func (c *Config) processQueue() {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for len(c.queue) > 0 {
			msg := c.queue[0]
			if c.writeToFile(msg.data) {
				c.queue = c.queue[1:]
				continue
			}
			if c.queue[0].retries >= c.NoOfRetries {
				c.queue = c.queue[1:]
			}
			c.queue[0].retries++

		}
	}()
	wg.Wait()
	c.isWritting = false

}

//Writre to file logic
func (c *Config) writeToFile(data string) bool {
	file := createFileIfNotExist("temp.txt")
	if getFileSize(file) > c.FileSize {
		file.Close()
		c.createTempFolderIfNotExist()
		c.moveFileToTemp()
		file = createFileIfNotExist("temp.txt")
	}
	status := appendToFile(file, data)
	file.Close()
	return status
}

func (c *Config) createTempFolderIfNotExist() {
	os.Mkdir(c.StoragePath, 0755)
}

func (c *Config) moveFileToTemp() { //Readfiles from here
	renameError := os.Rename("temp.txt", getFilePath(c.StoragePath))
	LogError("Cannot rename/move file", renameError)
}

func appendToFile(file *os.File, data string) bool {
	_, appendToFileError := file.WriteString(data)
	if appendToFileError != nil {
		LogError("Cannot append data to file", appendToFileError)
		return false
	} else {
		return true
	}
}

func createFileIfNotExist(fileName string) *os.File {
	file, fileOpenError := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	LogError("Cannot open file", fileOpenError)
	return file
}

func getFileSize(file *os.File) int64 {
	fileInfo, fileInfoError := file.Stat()
	if fileInfoError != nil {
		LogError("Unable to get file size", fileInfoError)
		return 0
	}
	const mb = 1024 * 1024
	return fileInfo.Size() / mb
}

func getFilePath(path string) string {
	return path + "/" + time.Now().String() + ".txt"
}

// Write to file logic end

//Read from file logic start
func (c *Config) Read(callback func(message string, fileName string)) {
	c.readFileAtTimeIntervel(30*time.Second, callback)
}

func (c *Config) readFileAtTimeIntervel(d time.Duration, callback func(message string, fileName string)) {
	for range time.Tick(d) {
		if !c.isReading {
			c.isReading = true
			var files = c.getAllFileFromDir()
			for _, file := range files {
				fileData := c.readFile(file.Name())
				fileInfo := convertFileDataToString(fileData)
				callback(fileInfo, file.Name())
			}
			c.isReading = false
		}
	}
}

//CommitFile :- removes file from storage
func (c *Config) CommitFile(fileName string) {
	c.removeFile(fileName)
}

func (c *Config) getAllFileFromDir() []os.FileInfo {
	files, filesInfoError := ioutil.ReadDir(c.StoragePath)
	LogError("Unable to get files form dir", filesInfoError)
	return files
}

func (c *Config) readFile(fileName string) []byte {
	fileData, fileReadError := ioutil.ReadFile(c.StoragePath + "/" + fileName) // just pass the file name
	LogError("Unable to read from file", fileReadError)
	return fileData
}

func convertFileDataToString(fileData []byte) string {
	return string(fileData)
}

func (c *Config) removeFile(fileName string) {
	removeFileError := os.Remove(c.StoragePath + "/" + fileName)
	LogError("Unable to delete file", removeFileError)
}

//LogError simple error log
func LogError(message string, errorData error) {
	if errorData != nil {
		LogError("Error from queue.go : -> "+message, errorData)
		return
	}
}
