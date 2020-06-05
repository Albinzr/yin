package socket

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// Config -
type Config struct {
	Network      string
	Address      string
	Listener     net.Listener
	OnConnect    func()
	OnDisconnect func()
}

//Init -
func (c *Config) Init() {
	ln, err := net.Listen(c.Network, c.Address)
	if err != nil {
		fmt.Println("error form socker init", err)
		return
	}
	defer ln.Close()
	c.Listener = ln
	//

	//

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("error form socker accept", err)
			return
		} else {
			//onConnect
		}

		netData, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		if strings.TrimSpace(string(netData)) == "STOP" {
			fmt.Println("Exiting TCP server!")
			return
		}

		fmt.Print("-> ", string(netData))
		t := time.Now()
		myTime := t.Format(time.RFC3339) + "\n"
		conn.Write([]byte(myTime))
	}

}
