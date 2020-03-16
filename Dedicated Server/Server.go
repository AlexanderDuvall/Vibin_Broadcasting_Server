package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

var synchronize = sync.Mutex{}
var wait = sync.WaitGroup{}
var broadcasters = make(map[int32][]byte)
var Broadcaster_ports = make([]bool, 20000, 20000)

type Message struct {
	Action int8
	Port   int32
	Error  int8
	key    string
}
type Server struct {
	ipAddress string
}
type Broadcaster struct {
	Address net.UDPAddr
	Key     string
	id      int32
	port    int32
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
		b.port = int32(port)
		address := net.UDPAddr{IP: net.ParseIP("localhost"), Port: address.Port}
		listener, err := net.ListenUDP("udp", &address)
		if err != nil {
			fmt.Println(err)
		}
		c := make(chan bool)
		confirmBroadcasterConnection(&b, *listener, c)
		defer listener.Close()
		proceed := <-c
		if proceed {
			err := listener.SetDeadline(time.Now().Add(5 * time.Second))
			for {
				//keep listening for song updates every three seconds
				if err != nil {
					fmt.Println(err)
				}
				time.Sleep(2 * time.Second)
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
func confirmBroadcasterConnection(broadcaster *Broadcaster, conn net.UDPConn, c chan bool) chan bool {
	a := createKey(20)
	var message = Message{1111, broadcaster.port, 0, a}
	m, err2 := json.Marshal(message)
	if err2 != nil {
		fmt.Println("unsuccessful: ", err2)
		c <- false
	}
	_, err := conn.WriteToUDP(m, &broadcaster.Address)
	if err != nil {
		fmt.Println("unsuccessful: ", err)
		c <- false
	}
	c <- true
	return c
}
func Prepare(conn net.UDPConn) {
	var bytes = make([]byte, 2048)
	//{4 digit code ,User_id,KeyIdentifier
	_, addr, _ := conn.ReadFromUDP(bytes)
	var request Request
	if err := json.Unmarshal(bytes, request); err != nil {
		fmt.Println(err)
	} else {
		a := request.Action
		switch a {
		case 211: //create broadcaster
			go newBroadcaster(*addr, request.id, request.key)
			break
		case 417: //send information
			break
		case 104: // remove broadcaster
			break
		}
	}
}
func startServer(server Server) {
	address := net.UDPAddr{IP: net.ParseIP(server.ipAddress), Port: 3000}
	listener, err := net.ListenUDP("udp", &address)
	defer listener.Close()
	if err != nil {
		fmt.Println(err)
	}
	defer listener.Close()
	for {
		time.Sleep(50 * time.Microsecond)
		go Prepare(*listener)
	}
}
func checkAndReservePorts() (found bool, port int) {
	wait.Add(1)
	synchronize.Lock()
	found = false
	port = -1
	for i, v := range Broadcaster_ports {
		if v == false {
			Broadcaster_ports[i] = true
			found = true
			port = i
			break
		}
	}
	defer synchronize.Unlock()
	defer wait.Done()
	return
}
func main() {
	wait.Wait()
	found, port := checkAndReservePorts()
	fmt.Println(found)
	fmt.Println(port)
}
