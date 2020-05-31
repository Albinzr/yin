package main

import (
	"runtime"

	server "applytics.in/yin/src"
)

func main() {
	runtime.GOMAXPROCS(2)
	server.Start()
}
