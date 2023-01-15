package main

import (
	"fmt"
	"log"

	//"os"
	"strconv"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	//"github.com/horizonledger/protocol"
)

func TestBasic(t *testing.T) {

	//cfile := "config.json"
	//config := getConfig(cfile)

	//TODO startup node and have all state available

	nodePort := 8000
	config := Config{Port: nodePort}

	log.Println(config)
	startupNodeStub(config)

	nodestate = NodeState{isLeader: true, msgstate: msgstate, vertexs: make(map[uuid.UUID]Vertex)}
	nodestate.unames = make(map[string]uuid.UUID)

	//go serveAll(config)

	time.Sleep(1 * time.Second)

	address := "ws://127.0.0.1:" + strconv.Itoa(nodePort) + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(address, nil)
	//msg := protocol.Msg{Type: "test", Value: "test"}

	if err != nil {
		fmt.Println("Cannot connect to websocket: ", address)
		t.Error("error websocket connect")
	}

	//log.Println("ws ", ws)
	sendMsg("test", ws)

	//TODO waiting
	// for _, v := range nodestate.vertexs {
	// 	log.Println(v.vertexid)
	// 	msg := <-v.in_read
	// 	log.Println(msg)
	// 	//protocol.Parse(msg)
	// 	if msg.Type != "Msg" {
	// 		t.Error("wrong type")
	// 	}
	// }

	time.Sleep(1 * time.Second)
	log.Println("....")
	log.Println(" ", len(nodestate.msgstate.MsgHistory))

	//os.Exit(1)
	//return

	// if err := ws.WriteMessage(
	// 	websocket.TextMessage,
	// 	protocol.ParseMessageToBytes(msg),
	// ); err != nil {
	// 	t.Error("error websocket connect")
	// 	//fmt.Println("WebSocket Write Error")
	// }

	// _, p, err := ws.ReadMessage()
	// if err != nil {
	// 	log.Println(err)
	// 	t.Error("error read ", err)
	// 	return
	// }

	// if string(p) != "test" {
	// 	t.Error("error read ", p)
	// }

	// inmsg := <-vertex.in_read

	// ntchan := ConnNtchanStub("test", "testout")
	// if ntchan.SrcName != "test" {
	// 	t.Error("setup error")
	// }

	// go func() { ntchan.REQ_in <- "test" }()

	// readout := <-ntchan.REQ_in

	// if readout != "test" {
	// 	t.Error("parse error")
	// }

}
