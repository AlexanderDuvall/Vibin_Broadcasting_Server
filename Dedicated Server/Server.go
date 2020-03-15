package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

type Server struct {
	ipAddress string
}
type Broadcaster struct {
	Address string
	Port    int
	Key     string
	id      int32
}
type broadcasterRequest struct {
	key      string
	songByte []byte
}
type exists interface {
	validate() bool
}

var broadcasters = make(map[int32][]byte)

func (b Broadcaster) validate() (exists bool) {
	exists = true
	//mysql function to determine validity---assume true for now
	return
}
func newBroadcaster(bytes []byte) {
	var b Broadcaster
	if err := json.Unmarshal(bytes, b); err != nil {
		fmt.Println(err)
	} else {
		if b.validate() {
			address := net.UDPAddr{IP: net.ParseIP(b.Address), Port: b.Port}
			listener, err := net.ListenUDP("udp", &address)
			if err != nil {
				fmt.Println(err)
			}
			defer listener.Close()
			for {
				//keep listening for song updates every three seconds
				time.Sleep(3 * time.Second)
				var bytes []byte
				for {
					//read bytes....
					var tempBytes = make([]byte, 8)
					n, err := listener.Read(tempBytes)
					if err != nil {
						fmt.Println()
						break
					}
					if (n == 0) {
						break
					}
					bytes = append(bytes, tempBytes...)
				}

				if len(bytes) == 0 {
					var request broadcasterRequest
					err = json.Unmarshal(bytes, request)
					if err != nil {
						fmt.Println(err)
					}
					if strings.Compare(b.Key, request.key) == 0 {
						//proceed
						fmt.Println(request.songByte)
					} else {
						fmt.Println("More than one entity attempted to access data...")
						break
					}
				}
			}
		}
	}
}

/**
Format Json Information and send it off to a new broadcaster
*/
func PrepareBroadcaster(conn net.Conn) {
	var bytes []byte
	//{4 digit code ,User_id,KeyIdentifier
	for {
		tempBytes := make([]byte, 8)
		n, _ := conn.Read(tempBytes)
		if (n == 0) {
			go newBroadcaster(bytes)
			break
		}
		bytes = append(bytes, tempBytes...)
	}
}

func startServer(server Server) {
	listener, err := net.Listen("udp", server.ipAddress)
	if err != nil {
		fmt.Println(err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if (err != nil) {
			fmt.Println(err)
		}
		go PrepareBroadcaster(conn)
	}
}

func main() {

}
