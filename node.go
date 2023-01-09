package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/horizonledger/protocol"
)

var state State

var stateFile = "state.json"

type State struct {
	LastUpdate time.Time      `json:"lastUpdate"`
	MsgHistory []protocol.Msg `json:"MsgHistory"`
	isLeader   bool
	vertexs    map[uuid.UUID]Vertex
}

type StateMsg struct {
	State State  `json:"state"`
	Type  string `json:"type"`
	Value string `json:"value"`
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
		state.vertexs[newuid] = vertex
		log.Println("OUTBOUND connection established")
		//set peer and connected state
	}

}

func readHandler(state *State, vertex *Vertex) {
	for {
		msg := <-(*vertex).in_read
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

	vertex.in_read = make(chan protocol.Msg)
	vertex.out_write = make(chan protocol.Msg)

	state.vertexs[newuid] = vertex

	//TODO only send/receive after handshake, for that we need to read only 1 message first and then open chans
	// --- handshake ---

	go readLoop(&vertex)
	go readHandler(&state, &vertex)
	go writeLoop(&vertex)

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
				xmsg := protocol.Msg{Type: "STATUS", Value: state.LastUpdate.String()}
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
			log.Println("#vertexs ", len(state.vertexs))
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

	go saveState(&state)
	go reportVertexs()

	state.vertexs = make(map[uuid.UUID]Vertex)
	log.Println("serve")
	go serveAll(config)

	port := "9000"
	host := "127.0.0.1"
	address := fmt.Sprintf("ws://%s:%s/ws", host, port)

	log.Println("connect outbound")
	connectOutbound(address)

	go statusLoop(state.vertexs)

}

// startup for testing
func startupNodeStub(config Config) {

	//TODO as args
	state.isLeader = true

	//TODO continously check leader election

	state.vertexs = make(map[uuid.UUID]Vertex)
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

func serveAll(config Config) {
	log.Println("serve on ", config.Port)
	setupRoutes()
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), nil))

}
