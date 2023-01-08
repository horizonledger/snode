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

var vertexs = map[uuid.UUID]Vertex{}
var state State

var stateFile = "state.json"

type State struct {
	LastUpdate time.Time `json:"lastUpdate"`
	MsgHistory []Msg     `json:"MsgHistory"`
	isLeader   bool
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

func connectOutbound(address string) {
	fmt.Println("connect outbound")
	//log.Fatal(http.ListenAndServe(":8000", nil))
	// address := url.URL{
	// 	Scheme: "ws",
	// 	Host:   "127.0.0.1",
	// 	Port:   8000,
	// }

	ws, _, err := websocket.DefaultDialer.Dial(address, nil)
	//fmt.Println("ws %T", ws)
	if err != nil {
		fmt.Println("Cannot connect to websocket: ", address)
	} else {
		fmt.Println("connected to websocket to ", address)
		sendMsg("HNDPEER", ws)

		var newuid = uuid.Must(uuid.NewV4())
		vertex := Vertex{wsConn: ws, vertexid: newuid, name: "default", handshake: false}
		vertexs[newuid] = vertex
		log.Println("OUTBOUND connection established")
		//set peer and connected state
	}

}

func readHandler(state State, vertex Vertex) {
	for {
		msg := <-vertex.in_read
		log.Println("readHandler.. ", msg)
		handleMsg(state, vertex, msg)
	}
}

// inbound ws connection
func serveWs(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	var newuid = uuid.Must(uuid.NewV4())
	vertex := Vertex{wsConn: ws, vertexid: newuid, name: "default", handshake: false}

	log.Println("INBOUND connection established")
	log.Println("vertex uuid ", newuid)

	vertex.in_read = make(chan Msg)
	vertex.out_write = make(chan Msg)

	vertexs[newuid] = vertex

	//TODO only send/receive after handshake, for that we need to read only 1 message first and then open chans
	// --- handshake ---

	go readLoop(vertex)
	go readHandler(state, vertex)
	go writeLoop(vertex)

	//wait for handshake from inbound
	//TODO timeout

	// //announce new client
	// for _, cl := range vertexs {
	// 	welcome_msg := Msg{Type: "chat", Value: newuid.String() + " entered"}
	// 	msgByte, _ := json.Marshal(welcome_msg)
	// 	fmt.Printf("send to %v %v\n", cl.vertexid, string(msgByte))
	// 	cl.wsConn.WriteMessage(1, msgByte)
	// }

}

func statusLoop(vertexs map[uuid.UUID]Vertex) {
	quit := make(chan struct{})
	ticker := time.NewTicker(20 * time.Second)
	for {
		//log.Println("statusLoop")
		select {
		case <-ticker.C:
			log.Println("status last update: ", state.LastUpdate.String())
			log.Println(len(vertexs))
			for _, v := range vertexs {
				xmsg := Msg{Type: "STATUS", Value: state.LastUpdate.String()}
				v.out_write <- xmsg

			}

		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func reportVertexs() {
	quit := make(chan struct{})
	ticker := time.NewTicker(20 * time.Second)
	for {
		select {
		case <-ticker.C:
			log.Println("#vertexs ", len(vertexs))
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func isLeader() bool {
	//TODO allocation of time slots depends on slotID
	//TODO get from config
	//if config.SlotID == 1
	if time.Now().Second() < 30 {
		return true
	} else {
		return false
	}
}

func handleMsg(state State, vertex Vertex, msg Msg) {

	log.Println("handle msg ", msg)
	log.Println("type >> ", msg.Type)

	switch msg.Type {
	case "STATUS":
		log.Println("STATUS received ")
		log.Println(">> ", msg.Value)
		//who is leader and follower?
		//if follower and new state then
		//pushState(vertex.wsConn)

	case "HNDPEER":
		//TODO check not already connected
		//TODO pubkey exchange here
		if vertex.handshake {
			log.Println("handle handshake already")
		} else {
			log.Println("handle handshake")
			xmsg := Msg{Type: "HNDSHAKEPEER", Value: "confirm"}
			vertex.out_write <- xmsg
			vertex.handshake = true
			vertex.isPeer = true
			vertex.isClient = false
		}

	case "HNDCLIENT":
		if vertex.handshake {
			log.Println("handle handshake already")
		} else {
			log.Println("handle handshake")
			xmsg := Msg{Type: "HNDSHAKECLIENT", Value: "confirm"}
			vertex.out_write <- xmsg
			vertex.handshake = true
			vertex.isPeer = false
			vertex.isClient = true

			pushState(vertex)

			//push uuid to client
			infoMsg := Msg{Type: "uuid", Value: vertex.vertexid.String()}
			vertex.out_write <- infoMsg

		}

	case "state":
		log.Println("handle state")

	case "chat":
		log.Println("handle chat")

		cid := vertex.vertexid.String()
		log.Println("vertex name: ", vertex.name)
		log.Println("vertexid: ", cid)
		if vertex.name != "default" {
			cid = vertex.name
		}
		textmsg := cid + ": " + msg.Value
		log.Println("textmsg ", textmsg)
		//broadcast
		fmt.Printf("vertexs len %v", len(vertexs))
		//TODO fix
		//broadcast(textmsg)

		//testing
		xmsg := Msg{Type: "chat", Value: textmsg}
		vertex.out_write <- xmsg

	case "name":
		//TODO check duplicate names on registration
		log.Println("handle name")
		fmt.Println("set name to: " + msg.Value)
		//TODO check if already registered
		//TODO name change doesnt persist
		vertex.name = msg.Value
		//state.
		fmt.Println("name now : " + vertex.name)

		xmsg := Msg{Type: "name", Value: msg.Value + "|registered"}
		vertex.out_write <- xmsg
		// msgByte, _ := json.Marshal(xmsg)
		// //cl.wsConn.WriteMessage(messageType, msgByte)
		// vertex.wsConn.WriteMessage(1, msgByte)
	default:
		return
	}
	//save state
	state.MsgHistory = append(state.MsgHistory, msg)

}

// TODO return node
func startupNode(config Config) {
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

	//TODO as args
	state.isLeader = true

	//TODO continously check leader election

	go saveState()
	go reportVertexs()

	vertexs = make(map[uuid.UUID]Vertex)
	log.Println("serve")
	go serveAll(config)

	port := "9000"
	host := "127.0.0.1"
	address := fmt.Sprintf("ws://%s:%s/ws", host, port)

	log.Println("connect outbound")
	connectOutbound(address)

	go statusLoop(vertexs)

}

// startup for testing
func startupNodeStub(config Config) {

	//TODO as args
	state.isLeader = true

	//TODO continously check leader election

	vertexs = make(map[uuid.UUID]Vertex)
	log.Println("serve")
	go serveAll(config)

	// port := "9000"
	// host := "127.0.0.1"
	// address := fmt.Sprintf("ws://%s:%s/ws", host, port)
	//log.Println("connect outbound")
	//connectOutbound(address)

}

func pushState(vertex Vertex) {
	statemsg := StateMsg{State: state, Type: "state"}
	//TODO fix message type in channel
	//vertex.out_write <- statemsg
	stateData, _ := json.MarshalIndent(statemsg, "", " ")
	_ = vertex.wsConn.WriteMessage(1, []byte(stateData))
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
	log.Println("serve on ", config.Port)
	setupRoutes()
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), nil))

}
