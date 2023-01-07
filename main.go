package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

var clients = map[uuid.UUID]Client{}
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

type Client struct {
	wsConn    *websocket.Conn
	clientid  uuid.UUID
	name      string
	in_read   chan (Msg)
	out_write chan (Msg)
	//channels
}

type Peer struct {
	wsConn   *websocket.Conn
	clientid uuid.UUID
	name     string
	//channels
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func handleMsg(client Client, msg Msg) {

	log.Println("handle msg ", msg)

	switch msg.Type {
	case "chat":
		log.Println("handle chat")

		cid := client.clientid.String()
		log.Println("client.name ", client.name)
		log.Println("cid ", cid)
		if client.name != "default" {
			cid = client.name
		}
		textmsg := cid + ": " + msg.Value
		//broadcast
		fmt.Printf("clients len %v", len(clients))
		broadcast(textmsg)

		//testing
		xmsg := Msg{Type: "chat", Value: textmsg}
		client.out_write <- xmsg

	case "name":
		log.Println("handle name")
		fmt.Print("name " + msg.Value)
		//TOOD check if already registered
		client.name = msg.Value

		xmsg := Msg{Type: "name", Value: msg.Value + "|registered"}
		client.out_write <- xmsg
		// msgByte, _ := json.Marshal(xmsg)
		// //cl.wsConn.WriteMessage(messageType, msgByte)
		// client.wsConn.WriteMessage(1, msgByte)
	default:
		return
	}
	//save state
	state.MsgHistory = append(state.MsgHistory, msg)

}

func readHandler(client Client) {
	for {
		msg := <-client.in_read
		handleMsg(client, msg)
	}
}

func pushState(ws *websocket.Conn) {
	statemsg := StateMsg{State: state, Type: "state"}
	stateData, _ := json.MarshalIndent(statemsg, "", " ")
	_ = ws.WriteMessage(1, []byte(stateData))
}

// inbound ws connection
func serveWs(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("ws connection")

	//TODO handshake in separate function

	//wait for handshake from inbound
	//TODO timeout

	_, inMsg, err := ws.ReadMessage()
	if err != nil {
		log.Println("err ", err)
	}
	log.Println("read ", string(inMsg))

	//determine client or peer
	switch string(inMsg) {
	case "HNDPEER":
		//check if already a peer?
		_ = ws.WriteMessage(1, []byte("ok"))

	case "HNDCLIENT":
		_ = ws.WriteMessage(1, []byte("HNDSRV"))
		pushState(ws)
	}

	//handle peer

	//handle client
	var newuid = uuid.Must(uuid.NewV4())
	client := Client{wsConn: ws, clientid: newuid, name: "default"}
	clients[newuid] = client
	log.Println("uid ", newuid)
	client.in_read = make(chan Msg)
	client.out_write = make(chan Msg)

	//reader(client)
	go readLoop(client)
	go readHandler(client)
	go writeLoop(client)

	client.out_write <- Msg{Type: "test", Value: "test"}

	//client.in_read <- Msg{Type: "test", Value: "test"}
	//client.in_read <- Msg{Type: "test", Value: "test"}

	// //announce new client
	// for _, cl := range clients {
	// 	welcome_msg := Msg{Type: "chat", Value: newuid.String() + " entered"}
	// 	msgByte, _ := json.Marshal(welcome_msg)
	// 	fmt.Printf("send to %v %v\n", cl.clientid, string(msgByte))
	// 	cl.wsConn.WriteMessage(1, msgByte)
	// }

	// //send uuid to client
	// infoMsg := Msg{Type: "uuid", Value: newuid.String()}
	// replyByte, _ := json.Marshal(infoMsg)

	// _ = ws.WriteMessage(1, replyByte)

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

func serveAll() {
	port := 8000
	fmt.Println("running on ", port)
	setupRoutes()
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))

	for {

	}
}

// TODO return node
func startupNode() {
	//check storage
	//var state State
	if !storageInited() {
		//if first start ever init storage
		state = initStorage()
		fmt.Println("state ", state)
	} else {
		fmt.Println("storage exists. last update...")
		state = loadStorage()
		fmt.Println(state.LastUpdate)
		updateSince := time.Since(state.LastUpdate)
		fmt.Println("updated ago ", updateSince)
		fmt.Println("messages stored ", len(state.MsgHistory))
	}

	//TODO routine to store state every x seconds
	// state.MsgHistory = append(state.MsgHistory, Msg{Type: "test", Value: "test"})
	// fmt.Println("state ", state)
	// writeState(state)

	go saveState()

	clients = make(map[uuid.UUID]Client)
	serveAll()
}

func main() {
	startupNode()
}
