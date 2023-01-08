package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

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

// inbound ws connection
func serveWs(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	var newuid = uuid.Must(uuid.NewV4())
	vertex := Vertex{wsConn: ws, vertexid: newuid, name: "default", handshake: false}
	vertexs[newuid] = vertex
	log.Println("INBOUND connection established")
	log.Println("vertex uuid ", newuid)

	vertex.in_read = make(chan Msg)
	vertex.out_write = make(chan Msg)

	//TODO only send/receive after handshake, for that we need to read only 1 message first and then open chans
	// --- handshake ---

	go readLoop(vertex)
	go readHandler(vertex)
	go writeLoop(vertex)

	// 	vertex.out_write <- Msg{Type: "test", Value: "test"}

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

func heartbeatLoop(vertexs map[uuid.UUID]Vertex) {
	log.Println("heartbeatLoop")
	quit := make(chan struct{})
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			log.Println("heartbeat..")
			for _, cl := range vertexs {
				xmsg := Msg{Type: "HEARTBEAT", Value: "..."}
				log.Println("write ", xmsg)
				cl.out_write <- xmsg
			}

		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func handleMsg(vertex Vertex, msg Msg) {

	log.Println("handle msg ", msg)

	switch msg.Type {
	case "HEARTBEAT":
		log.Println("HEARTBEAT received")

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

			pushState(vertex.wsConn)

			//push uuid to client
			infoMsg := Msg{Type: "uuid", Value: vertex.vertexid.String()}
			vertex.out_write <- infoMsg

		}

	case "chat":
		log.Println("handle chat")

		cid := vertex.vertexid.String()
		log.Println("client.name ", vertex.name)
		log.Println("cid ", cid)
		if vertex.name != "default" {
			cid = vertex.name
		}
		textmsg := cid + ": " + msg.Value
		//broadcast
		fmt.Printf("vertexs len %v", len(vertexs))
		broadcast(textmsg)

		//testing
		xmsg := Msg{Type: "chat", Value: textmsg}
		vertex.out_write <- xmsg

	case "name":
		log.Println("handle name")
		fmt.Print("name " + msg.Value)
		//TOOD check if already registered
		vertex.name = msg.Value

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

func reportVertexs() {
	quit := make(chan struct{})
	ticker := time.NewTicker(5 * time.Second)
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

	//TODO routine to store state every x seconds
	// state.MsgHistory = append(state.MsgHistory, Msg{Type: "test", Value: "test"})
	// fmt.Println("state ", state)
	// writeState(state)

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

	go heartbeatLoop(vertexs)

}
