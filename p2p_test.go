package main

import (
	"log"
	"strconv"
	"testing"

	"github.com/gorilla/websocket"
)

func TestBasic(t *testing.T) {

	go serveAll()
	port := 8000
	configStr := "ws://127.0.0.1:" + strconv.Itoa(port) + "/ws"
	ws, _, _ := websocket.DefaultDialer.Dial(configStr, nil)

	//log.Println(ws)

	sendMsg("test", ws)

	log.Println(clients)

	// if err := ws.WriteMessage(
	// 	websocket.TextMessage,
	// 	[]byte("test"),
	// ); err != nil {
	// 	fmt.Println("WebSocket Write Error")
	// }

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
