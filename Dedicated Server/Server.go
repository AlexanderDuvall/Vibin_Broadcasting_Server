package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var synchronize = sync.Mutex{}
var wait = sync.WaitGroup{}
var broadcasters = make(map[int32][]byte)
var BroadcasterPorts = make([]bool, 20000, 20000)
var ListenerPorts = make([]bool, 1, 2000)
var listenerOffset, broadcasterOffset = 1000, 3001

type Message struct {
	Action int16
	Port   int16
	Error  int8
	Key    string
}
type Server struct {
	Address net.UDPAddr
}
type Broadcaster struct {
	Address net.UDPAddr
	Key     string
	id      int32
}
type Request struct {
	Action    uint16 //4 digit
	key       string
	id        int32
	songBytes []byte
}

type exists interface {
	validate() bool
}

func (b Broadcaster) validate() (exists bool) {
	exists = true
	//mysql function to determine validity---assume true for now
	return
}

func newBroadcaster(address net.UDPAddr, id int32, key string) {
	var b Broadcaster
	b = Broadcaster{Address: address, Key: key, id: id}
	if found, port := checkAndReservePorts(); b.validate() && found {
		broadcasters[b.id] = nil
		port += broadcasterOffset
		b.Address.Port = port
		fmt.Println("....port: ", port)
		address := net.UDPAddr{IP: net.ParseIP("localhost"), Port: port}
		//create listener for broadcaster
		listener, err := net.ListenUDP("udp", &address)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("confirming broadcaster conn...")

		defer listener.Close()
		proceed := confirmBroadcasterConnection(&b)
		fmt.Println("confirmed")
		if proceed {
			deadline := time.Now().Add(5 * time.Second)
			for time.Now().Before(deadline) {
				//keep listening for song updates every three seconds
				if err != nil {
					fmt.Println(err)
				}
				time.Sleep(1500 * time.Millisecond)
				var bytes = make([]byte, 2048)
				//read bytes....
				_, err = listener.Read(bytes)
				if err != nil && strings.Compare(err.Error(), "EOF") != 0 {
					fmt.Println(err)
					break
				} else if len(bytes) != 0 {
					var request Request
					err = json.Unmarshal(bytes, request)
					if err != nil {
						fmt.Println(err)
					}
					//authenticate user
					if strings.Compare(b.Key, request.key) == 0 {
						err := listener.SetDeadline(time.Now().Add(5 * time.Second))
						if err != nil {
							fmt.Println(err)
						}
						//proceed
						wait.Add(1)
						synchronize.Lock()
						//lock broadcaster key at Id
						broadcasters[b.id] = request.songBytes
						synchronize.Unlock()
						wait.Done()
					} else {
						fmt.Println("More than one entity attempted to access data...")
						break
					}
				}
			}
		} else {
			fmt.Println("failure")
		}
	} else {
		fmt.Println("Failure...aa")
	}
}
func createKey(length int8) string {
	letter := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

/**
Format Json Information and send it off to a new broadcaster
*/
func confirmBroadcasterConnection(broadcaster *Broadcaster) bool {
	a := createKey(20)
	var message = Message{1111, int16(broadcaster.Address.Port), 0, a}
	m, err2 := json.Marshal(message)
	if err2 != nil {
		fmt.Println("unsuccessful: ", err2)
		return false
	}

	fmt.Println(broadcaster.Address.String())
	fmt.Println("dialing")
	conn, err := net.Dial("udp", ":4411")
	if err != nil {
		fmt.Println("unsuccessful: ", err)
		return false
	}
	fmt.Println("sending..")
	fmt.Println(m)
	_, err = conn.Write(m)
	fmt.Println("sent")
	fmt.Println("ending")
	return true
}
func Prepare(conn net.UDPConn) {
	var bytes = make([]byte, 2048)
	//{4 digit code ,User_id,KeyIdentifier
	amt, address, err2 := conn.ReadFromUDP(bytes)
	fmt.Println("address read from...", address.Port)
	var request Request
	err := json.Unmarshal(bytes[:amt], &request)
	fmt.Println(request)
	if err != nil || err2 != nil {
		fmt.Println(err)
		fmt.Println(err2)
	} else {
		a := request.Action
		switch a {
		case 2111: //create broadcaster
			fmt.Println("creating new broadcaster")
			newBroadcaster(*address, request.id, request.key)
			break
		case 417: //send information
			break
		case 104: // remove broadcaster
			break
		}
	}
}
func requestBroadcaster(w http.ResponseWriter, r *http.Request) {

}
func startWebServer(w http.ResponseWriter, r *http.Request) {
	http.HandleFunc("/request", requestBroadcaster)
}
func startServer(server Server) {
	address := net.UDPAddr{IP: server.Address.IP, Port: server.Address.Port}
	listener, err := net.ListenUDP("udp", &address)
	if err != nil {
		fmt.Println(err)
	}
	defer listener.Close()
	fmt.Println("Starting Server....")
	for {
		fmt.Println("Waiting...")
		time.Sleep(50 * time.Microsecond)
		go Prepare(*listener)
	}
}
func checkAndReservePorts() (found bool, port int) {
	fmt.Println("checking and reserving")
	wait.Add(1)
	synchronize.Lock()
	defer synchronize.Unlock()
	defer wait.Done()
	found = false
	port = -1
	for i, v := range BroadcasterPorts {
		if v == false {
			BroadcasterPorts[i] = true
			found = true
			port = i
			break
		}
	}

	return
}
func main() {
	for i := 0; i <= len(ListenerPorts); i++ {
		fmt.Println("ffff")
		conn := net.UDPAddr{IP: net.ParseIP("aa"), Port: i + listenerOffset}
		fmt.Println(conn.Port)
		server := Server{Address: conn}
		startServer(server)
	}
}
