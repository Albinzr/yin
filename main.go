package main

import (
	server "applytics.in/yin/src"
)

func main() {
	// runtime.GOMAXPROCS(4)
	server.Start()
}
