package main

import (
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
	//TODO change entry pubkeys
	unames      map[string]uuid.UUID
	cashbalance map[string]uint64
	msgstate    MsgState
	vertexs     map[uuid.UUID]Vertex
	pubsub      *Pubsub
}

type MsgState struct {
	LastUpdate time.Time      `json:"lastUpdate"`
	MsgHistory []protocol.Msg `json:"MsgHistory"`
}

func syncHistory(vertex *Vertex) {
	if len(nodestate.msgstate.MsgHistory) == 0 {
		log.Info("our height is 0.. pull all")
		xmsg := protocol.Msg{Type: REQSTATE, Value: strconv.Itoa(0)}
		vertex.out_write <- protocol.MsgToGen(xmsg)
	} else {
		n := len(nodestate.msgstate.MsgHistory)
		log.Info("n ", n)
		lastMsgTime := nodestate.msgstate.MsgHistory[n-1].Time
		//TODO need to compare duration and figure out if we're behind
		log.Info("our last time : ", lastMsgTime)
	}
}

func connectOutbound(address string) {
	log.Debug("connect outbound")

	ws, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		log.Warn("Cannot connect to websocket: ", address)
	} else {
		log.Info("connected to websocket to ", address)

		var newuid = uuid.Must(uuid.NewV4())
		vertex := Vertex{wsConn: ws, vertexid: newuid, name: "default", handshake: false}
		vertex.in_read = make(chan protocol.Gen)
		vertex.out_write = make(chan protocol.Gen)
		nodestate.vertexs[newuid] = vertex
		log.Info("OUTBOUND connection established")

		go readLoop(&vertex)
		go readHandler(&nodestate, &vertex)
		go writeLoop(&nodestate, &vertex)

		msg := protocol.Msg{Type: "HNDPEER", Value: "REQUEST"}
		log.Debug("send ", msg)
		vertex.out_write <- protocol.MsgToGen(msg)
		//set peer and connected state
	}

}

func subToOut(vertex *Vertex, genchan chan protocol.Gen) {
	for {
		log.Println("waiting ")
		//TODO put in
		msg := <-genchan
		vertex.out_write <- msg
		log.Println("..sub read ", msg)
	}
}

func hookVertexToSub(vertex *Vertex, topic string) {
	ch1 := make(chan protocol.Gen)
	go subToOut(vertex, ch1)
	nodestate.pubsub.Subscribe(topic, ch1)
}

// inbound ws connection
func serveWs(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("error ", err)
	}

	var newuid = uuid.Must(uuid.NewV4())
	vertex := Vertex{wsConn: ws, vertexid: newuid, name: "default", handshake: false}

	log.Info("INBOUND connection established")
	log.Info("vertex uuid ", newuid)

	vertex.in_read = make(chan protocol.Gen)
	vertex.out_write = make(chan protocol.Gen)

	nodestate.vertexs[newuid] = vertex

	//TODO only send/receive after handshake, for that we need to read only 1 message first and then open chans
	// --- handshake ---

	log.Info("init readLoop")
	go readLoop(&vertex)
	log.Info("init readHandler")
	go readHandler(&nodestate, &vertex)
	log.Info("init writeLoop")
	go writeLoop(&nodestate, &vertex)

	//TODO do this on SUB message from subscriber
	//hookVertexToSub(&vertex, "vertex")
	//hookVertexToSub(&vertex, "status")

	//wait for handshake from inbound
	//TODO timeout

	// TODO //announce new client
	// for _, cl := range vertexs {
	// 	welcome_msg := Msg{Type: "chat", Value: newuid.String() + " entered"}
	// 	msgByte, _ := json.Marshal(welcome_msg)
	// 	fmt.Printf("send to %v %v\n", cl.vertexid, string(msgByte))
	// 	cl.wsConn.WriteMessage(1, msgByte)
	// }

}

func syncLoop(vertexs map[uuid.UUID]Vertex) {
	//for each vertex check the height

	//if we're behind then
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
				//only sync with peers not clients
				if v.isPeer {
					xmsg := protocol.Msg{Type: "REQHEIGHT", Value: ""}
					log.Info("request height: ", xmsg)
					v.out_write <- protocol.MsgToGen(xmsg)
				}
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

func initGenesis(nodestate NodeState) {
	//read from json file
	nodestate.cashbalance["satoshi"] = 200
	nodestate.cashbalance["hal"] = 100
}

func applyCashTx(nodestate NodeState, cashtx protocol.CashTx) {
	//func transferCash(nodestate NodeState, from string, to string, amount uint64) {
	//checks
	nodestate.cashbalance[cashtx.Sender] -= cashtx.Amount
	nodestate.cashbalance[cashtx.Receiver] += cashtx.Amount
}

func showAllBal(nodestate NodeState) {
	for r := range nodestate.cashbalance {
		log.Info(r, " has ", nodestate.cashbalance[r])
	}
}

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
		n := len(msgstate.MsgHistory)
		if n > 0 {
			log.Info("last message  ", msgstate.MsgHistory[n-1])
			log.Info("last message time ", msgstate.MsgHistory[n-1].Time)
		}
	}

	//currently set statically once
	if config.NodeID == 1 {
		nodestate.isLeader = true
	} else {
		nodestate.isLeader = false
	}

	//TODO continously check leader election

	go saveState(config.StateFile, &nodestate)

	go syncState()

	nodestate = NodeState{isLeader: true, msgstate: msgstate, vertexs: make(map[uuid.UUID]Vertex)}
	nodestate.unames = make(map[string]uuid.UUID)
	nodestate.cashbalance = make(map[string]uint64)

	initGenesis(nodestate)

	//TESTING
	showAllBal(nodestate)
	tx := protocol.CashTx{Sender: "satoshi", Receiver: "hal", Amount: 50}
	applyCashTx(nodestate, tx)
	showAllBal(nodestate)

	nodestate.pubsub = NewPubsub()

	log.Info("serve")
	go serveAll(config)

	for _, v := range config.InitVertex {
		//address := fmt.Sprintf("ws://%s:%s/ws", host, port)
		address := fmt.Sprintf("ws://%s/ws", v)

		log.Info("connect outbound to ", address)
		connectOutbound(address)
	}

	//start publishing default topics
	go pubVertexs(nodestate.pubsub)
	go pubStatus(nodestate.pubsub)

}

func sendMsg(textmsg string, ws *websocket.Conn) {
	msg := protocol.Msg{Type: "Msg", Value: textmsg}
	gen := protocol.MsgToGen(msg)

	log.Info("send ", gen)
	if err := ws.WriteMessage(
		websocket.TextMessage,
		protocol.ParseGenToBytes(gen),
	); err != nil {
		fmt.Println("WebSocket Write Error")
	}
}

func serveAll(config Config) {
	log.Info("serve on ", config.Port)
	setupRoutes()
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), nil))

}

// startup for testing
func startupNodeStub(config Config) {

	nodestate = NodeState{isLeader: true, msgstate: msgstate, vertexs: make(map[uuid.UUID]Vertex)}
	nodestate.unames = make(map[string]uuid.UUID)
	var newuid = uuid.Must(uuid.NewV4())
	nodestate.unames["test"] = newuid

	log.Info("serve")
	go serveAll(config)

}
