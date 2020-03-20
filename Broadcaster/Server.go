package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

var server Server
var headers = []string{"HTTP/1.1 200 OK\n",
	"Date: Tue, 14 Dec 2010 10:48:45 GMT\n",
	"Server: GoLang\n",
	"Content-Type: text/html; charset=iso-8859-1\n",
	"Content-Length: 4\n", "Machine-Reached-Status: true\n"}

type Response struct {
	Status               string
	Date                 time.Time
	ContentType          string
	ContentLength        int16
	MachineReachedStatus bool
}
type Server struct {
	Address net.UDPAddr
	Key     string
	port    int16
}
type Request struct {
	Action    int16
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

func beginBroadcasting(stop chan bool) chan bool {
	deadline := time.Now().Add(5 * time.Minute)
	address := net.UDPAddr{IP: net.ParseIP("localhost"), Port: 4447}
	c, err := net.ListenUDP("udp", &address)
	CheckforErrors(err)
	for time.Now().Before(deadline) {
		bytes := make([]byte, 2048)
		time.Sleep(1500 * time.Microsecond)
		//attempting local contact...
		_, address, err := c.ReadFromUDP(bytes)
		var request Request
		proceed := checkAddress(address)
		if proceed {
			deadline = time.Now().Add(6 * time.Second)
			sendOff(request)
			err = json.Unmarshal(bytes, request)
			CheckforErrors(err)
		}
	}
	stop <- true
	return stop
}

func localRequest() {

}
func startServer() {

	ch := make(chan bool)
	listen, _ := net.Listen("tcp", ":4444")
	for {
		fmt.Println("Waiting to make contact")
		bytes := make([]byte, 2048)
		time.Sleep(2 * time.Second)
		//attempting local contact...
		conn, err := listen.Accept()

		var request Request
		_, err = conn.Read(bytes)
		CheckforErrors(err)
		fmt.Println(string(bytes))
		err = json.Unmarshal(bytes, request)
		CheckforErrors(err)
		//connection set between vibin and localhost
		proceed := checkAddress(conn.LocalAddr())
		if proceed {
			switch request.Action {
			case 1001: //begin first connection
				confirmreach(conn, ch) //todo finish function please
				x := <-ch
				if x {
					ch <- false
					break
				}
				_ = listen.Close()
				os.Exit(1)
				server.port, server.Key = contactRemoteServer()
				_ = conn.Close()
				beginBroadcasting(ch)
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
func confirmreach(conn net.Conn, c chan bool) {
	fmt.Println("Machine Reached...")

	var bytes []byte
	bytes, _ = json.Marshal(headers)
	_, err := conn.Write(bytes)
	if err != nil {
		c <- true
	}
	//todo update UI
}

func checkAddress(address net.Addr) bool {
	// will compare strings
	name, err := net.LookupAddr(address.String())
	fmt.Println(name, err)
	return true
}
func handleFirstConnection(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	bytes, err := json.Marshal(Response{Status: "200", Date: time.Now(), ContentLength: 4, ContentType: "text/html; charset=iso-8859-1\n", MachineReachedStatus: true})
	w.Header().Set("Access-Control-Allow-Origin", "*") //todo replace when out of testing
	_, err = w.Write(bytes)
	if err != nil {
		panic(err)
	}
}
func handleSongBytes(w http.ResponseWriter, r *http.Request) {
}
func startWebServer() {
	fmt.Println("Preparing Web Server")
	http.HandleFunc("/establish", handleFirstConnection)
	http.HandleFunc("/songBytes", handleSongBytes)
	_ = http.ListenAndServe(":4447", nil)
	fmt.Println("Web Server has started")
}
func main() {
	startWebServer()
	//	startServer()
}
