package server

import (
	"net/url"
	"testing"

	util "applytics.in/yin/src/helpers"
	"github.com/googollee/go-engine.io/transport/websocket"
)

func Test(t *testing.T) {
	// main()
	createClient()
}

func createClient() {

	transporter := websocket.Default

	u, err := url.Parse("http://localhost:8080")
	if err != nil {
		util.LogFatal("cannot parse url")
	}
	transporter.Dial(u, nil)
	transporter.Accept(nil, nil)

}
