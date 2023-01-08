package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {

	//cfile := "config.json"
	//config := getConfig(cfile)

	//TODO startup node and have all state available

	config := Config{Port: 8000}

	log.Println(config)
	go serveAll(config)

	time.Sleep(1 * time.Second)

	port := 8000
	address := "ws://127.0.0.1:" + strconv.Itoa(port) + "/ws"
	ws, _, err := websocket.DefaultDialer.Dial(address, nil)
	if err != nil {
		fmt.Println("Cannot connect to websocket: ", address)
		return
	} else {
		//log.Println("ws ", ws)
		sendMsg("test", ws)

	}

	log.Println(vertexs)

	if err := ws.WriteMessage(
		websocket.TextMessage,
		[]byte("test"),
	); err != nil {
		fmt.Println("WebSocket Write Error")
	}

	// inmsg := <-vertex.in_read
	// fmt.Println("insg ", inmsg)

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
