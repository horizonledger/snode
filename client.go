package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

func sendMsg(msg string, ws *websocket.Conn) {

	fmt.Println("send ", msg)
	if err := ws.WriteMessage(
		websocket.TextMessage,
		[]byte(msg),
	); err != nil {
		fmt.Println("WebSocket Write Error")
	}
}

//advanced handshake
//send challenge
//verify sig

func runclient() {
	fmt.Println("running on 8000")
	//log.Fatal(http.ListenAndServe(":8000", nil))

	configStr := "ws://127.0.0.1:8000/ws"
	ws, _, err := websocket.DefaultDialer.Dial(configStr, nil)
	fmt.Printf("ws %T", ws)
	if err != nil {
		fmt.Println("Cannot connect to websocket: ", configStr)
	} else {
		fmt.Println("connected to websocket to ", configStr)
	}

	sendMsg("HNDPEER", ws)

	//wait for handshake reply
	// for {
	// 	_, msg, err := ws.ReadMessage()
	// 	if err != nil {
	// 		//conn.log("listen", err, "Cannot read websocket message")
	// 		//conn.closeWs()
	// 		break
	// 	}
	// 	fmt.Println("received msg: ", string(msg))

	// 	if string(msg) == "ok" {
	// 		fmt.Println("hanshake complete")
	// 	}
	// }

	//fmt.Println(ws)

}
