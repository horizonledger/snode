package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

type Config struct {
	NodeAlias string
	Verbose   bool
	NodePort  int
	WebPort   int
	// CreateGenesis bool
}

var vertexs = map[uuid.UUID]Vertex{}
var state State

var stateFile = "state.json"

type State struct {
	LastUpdate time.Time `json:"lastUpdate"`
	MsgHistory []Msg     `json:"MsgHistory"`
}

type StateMsg struct {
	State State  `json:"state"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Msg struct {
	Type   string    `json:"type"`
	Value  string    `json:"value"`
	Sender uuid.UUID `json:"uuid,omitempty"`
	Time   time.Time `json:"time,omitempty"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func readHandler(vertex Vertex) {
	for {
		msg := <-vertex.in_read
		handleMsg(vertex, msg)
	}
}

func pushState(ws *websocket.Conn) {
	statemsg := StateMsg{State: state, Type: "state"}
	stateData, _ := json.MarshalIndent(statemsg, "", " ")
	_ = ws.WriteMessage(1, []byte(stateData))
}

func getConfig() Config {

	conffile := "config.json"
	log.Println("config file ", conffile)

	if _, err := os.Stat(conffile); os.IsNotExist(err) {
		log.Println("config file does not exist. create a file named ", conffile)
		//return nil
	}

	content, err := ioutil.ReadFile(conffile)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	return config

}

func setupRoutes() {
	// fs := http.FileServer(http.Dir("./static"))
	// http.Handle("/", fs)

	http.HandleFunc("/ws", serveWs)
}

func writeState(state State) {
	//log.Println("write state ", state)
	jsonData, _ := json.MarshalIndent(state, "", " ")
	//fmt.Println(string(jsonData))
	_ = ioutil.WriteFile(stateFile, jsonData, 0644)
}

func initStorage() State {
	fmt.Println("init storage")
	var emptyHistory []Msg
	state := State{LastUpdate: time.Now(), MsgHistory: emptyHistory}
	writeState(state)
	return state
}

func storageInited() bool {
	if _, err := os.Stat(stateFile); err == nil {
		return true
	} else {
		return false
	}
}

func loadStorage() State {

	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}
	state := State{}
	if err := json.Unmarshal(data, &state); err != nil {
		panic(err)
	}
	return state
}

func saveState() {
	//TODO store only if state has changed
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			writeState(state)
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func serveAll(config Config) {

	fmt.Println("running on ", config.WebPort)
	setupRoutes()
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.WebPort), nil))

}

// type fu func()

// func doContinous(f fu) {
// 	quit := make(chan struct{})
// 	ticker := time.NewTicker(5 * time.Second)
// 	for {
// 		select {
// 		case <-ticker.C:
// 			//
// 			f()
// 		case <-quit:
// 			ticker.Stop()
// 			return
// 		}
// 	}
// }

var (
	//env  *string
	port       *int
	configFile *string
)

func init() {
	//env = flag.String("env", "development", "current environment")
	port = flag.Int("port", 8000, "port number")
	configFile = flag.String("config", "", "port number")
}

func main() {
	flag.Parse()
	log.Println("?? ", *port)
	//config := getConfig()
	config := Config{WebPort: *port}
	log.Println(config)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	startupNode(config)
	<-quit // This will block until you manually exists with CRl-C
	log.Println("\nnode exiting")
}
