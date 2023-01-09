package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/horizonledger/protocol"
)

var msgstate MsgState
var nodestate NodeState

type NodeState struct {
	isLeader bool
	msgstate MsgState
	vertexs  map[uuid.UUID]Vertex
}

type MsgState struct {
	LastUpdate time.Time      `json:"lastUpdate"`
	MsgHistory []protocol.Msg `json:"MsgHistory"`
}

func connectOutbound(address string) {
	log.Debug("connect outbound")
	//log.Fatal(http.ListenAndServe(":8000", nil))
	// address := url.URL{
	// 	Scheme: "ws",
	// 	Host:   "127.0.0.1",
	// 	Port:   8000,
	// }

	ws, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		log.Warn("Cannot connect to websocket: ", address)
	} else {
		log.Info("connected to websocket to ", address)
		//sendMsg("HNDPEER", ws)

		var newuid = uuid.Must(uuid.NewV4())
		vertex := Vertex{wsConn: ws, vertexid: newuid, name: "default", handshake: false}
		//TODO channels
		vertex.in_read = make(chan protocol.Msg)
		vertex.out_write = make(chan protocol.Msg)
		nodestate.vertexs[newuid] = vertex
		log.Info("OUTBOUND connection established")

		go readLoop(&vertex)
		go readHandler(&nodestate, &vertex)
		go writeLoop(&vertex)

		msg := protocol.Msg{Type: "HNDPEER", Value: "REQUEST"}
		log.Debug("send ", msg)
		vertex.out_write <- msg
		//set peer and connected state
	}

}

func readHandler(nodestate *NodeState, vertex *Vertex) {
	for {
		msg := <-(*vertex).in_read
		log.Debug("readHandler.. ", msg)
		handleMsg(nodestate, vertex, msg)
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

	log.Info("INBOUND connection established")
	log.Info("vertex uuid ", newuid)

	vertex.in_read = make(chan protocol.Msg)
	vertex.out_write = make(chan protocol.Msg)

	nodestate.vertexs[newuid] = vertex

	//TODO only send/receive after handshake, for that we need to read only 1 message first and then open chans
	// --- handshake ---

	go readLoop(&vertex)
	go readHandler(&nodestate, &vertex)
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
			log.Debug("status last update: ", nodestate.msgstate.LastUpdate.String())
			log.Debug(len(vertexs))
			for _, v := range vertexs {
				xmsg := protocol.Msg{Type: "STATUS", Value: nodestate.msgstate.LastUpdate.String()}
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
			log.Debug("#vertexs ", len(nodestate.vertexs))
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

// continously query peers
func syncState() {

	quit := make(chan struct{})
	ticker := time.NewTicker(20 * time.Second)
	for {
		select {
		case <-ticker.C:
			log.Info("query state height ")
			for _, v := range nodestate.vertexs {
				xmsg := protocol.Msg{Type: "REQHEIGHT", Value: ""}
				log.Info("request height: ", xmsg)
				v.out_write <- xmsg
			}
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
	if !storageInited(config.StateFile) {
		//if first start ever init storage
		msgstate = initStorage(config.StateFile)
		log.Debug("state ", msgstate)
	} else {
		log.Info("storage exists. last update...")
		msgstate = loadStorage(config.StateFile)
		log.Debug(msgstate.LastUpdate)
		updateSince := time.Since(msgstate.LastUpdate)
		log.Info("updated ago ", updateSince)
		log.Info("messages stored ", len(msgstate.MsgHistory))
	}

	//currently set statically once
	if config.NodeID == 1 {
		nodestate.isLeader = true
	} else {
		nodestate.isLeader = false
	}

	//TODO continously check leader election

	go saveState(config.StateFile, &msgstate)
	go reportVertexs()
	go syncState()

	//TODO
	nodestate = NodeState{isLeader: true, msgstate: msgstate, vertexs: make(map[uuid.UUID]Vertex)}

	log.Info("serve")
	go serveAll(config)

	for _, v := range config.InitVertex {
		// port := "9000"
		// host := "127.0.0.1"
		//address := fmt.Sprintf("ws://%s:%s/ws", host, port)
		address := fmt.Sprintf("ws://%s/ws", v)

		log.Info("connect outbound to ", address)
		connectOutbound(address)
	}

	go statusLoop(nodestate.vertexs)

}

func sendMsg(msg string, ws *websocket.Conn) {

	log.Debug("send ", msg)
	if err := ws.WriteMessage(
		websocket.TextMessage,
		[]byte(msg),
	); err != nil {
		fmt.Println("WebSocket Write Error")
	}
}

// startup for testing
// func startupNodeStub(config Config) {

// }

func pushState(vertex Vertex) {
	msgData, _ := json.MarshalIndent(nodestate.msgstate.MsgHistory, "", " ")

	statemsg := protocol.Msg{Type: "state", Value: string(msgData)}
	//TODO fix message type in channel
	//vertex.out_write <- statemsg
	stateData, _ := json.MarshalIndent(statemsg, "", " ")
	_ = vertex.wsConn.WriteMessage(1, []byte(stateData))
}

func serveAll(config Config) {
	log.Info("serve on ", config.Port)
	setupRoutes()
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), nil))

}
