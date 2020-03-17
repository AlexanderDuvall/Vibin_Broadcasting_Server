package Broadcaster

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

var server Server

type Server struct {
	Address net.UDPAddr
	Key     string
	port    int16
}
type Request struct {
	Action    int8
	Key       string
	Id        int32
	SongBytes []byte
	Port      int16
}
type Message struct {
	Action int16
	Port   int16
	Error  int8
	Key    string
}

func contactRemoteServer() (port int16, key string) {
	conn, err := net.Dial("udp", "remotehost:port") //todo change address
	CheckforErrors(err)
	message := Request{Action: 1001, Key: "", Id: 2} //first time connection, port not needed
	bytes, err := json.Marshal(message)
	//send info off
	_, err = conn.Write(bytes)
	CheckforErrors(err)
	address := net.UDPAddr{IP: net.ParseIP(""), Port: 2222}
	//listen for update on request
	listener, err := net.ListenUDP("udp", &address)
	CheckforErrors(err)
	deadline := time.Now().Add(10 * time.Second)
	bytes = make([]byte, 2048)
	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)
		_, _, err := listener.ReadFromUDP(bytes)
		CheckforErrors(err)
		var m Message
		_ = json.Unmarshal(bytes, m)
		if m.Error == 0 {
			if m.Action == 1111 {
				//GOOD TO GO
				port = m.Port
				key = m.Key
				return
			}
		}
	}
	return
}
func sendOff(request Request) {
	conn, err := net.Dial("udp", "remotehost:port") //todo change address
	CheckforErrors(err)
	message, err := json.Marshal(request) //first time connection, port not needed
	//send info off
	_, err = conn.Write(message)
	CheckforErrors(err)
}

func beginBroadcasting(conn net.UDPConn, stop chan bool) chan bool {
	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) {
		bytes := make([]byte, 2048)
		time.Sleep(1500 * time.Microsecond)
		//attempting local contact...
		_, address, err := conn.ReadFromUDP(bytes)
		var request Request
		err = json.Unmarshal(bytes, request)
		CheckforErrors(err)
		proceed := checkAddress(*address)
		if proceed {
			deadline = time.Now().Add(6 * time.Second)
			sendOff(request)
		}
	}
	stop <- true
	return stop
}
func startServer() {
	conn, err := net.ListenUDP("udp", &server.Address)
	if err != nil {
		fmt.Println("err")
	}
	ch := make(chan bool)
	for {
		CheckforErrors(err)
		bytes := make([]byte, 2048)
		time.Sleep(2 * time.Second)
		//attempting local contact...
		_, address, err := conn.ReadFromUDP(bytes)
		var request Request
		err = json.Unmarshal(bytes, request)
		CheckforErrors(err)
		//connection set between vibin and localhost
		proceed := checkAddress(*address)
		if proceed {
			switch request.Action {
			case 1001: //begin first connection
				server.port, server.Key = contactRemoteServer()
				beginBroadcasting(*conn, ch)
				<-ch
				fmt.Println("restarting")
				//begin listening for
				ch <- false
				break
			case 1002:
				break
			}
		}
	}
}
func checkAddress(address net.UDPAddr) bool {
	name, err := net.LookupAddr(address.IP.String())
	fmt.Println(name, err)
	return true
}
func main() {
	address := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 4447}
	server = Server{Address: address}
	startServer()
}
