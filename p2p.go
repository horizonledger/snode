package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

// Vertex is a wrapper around a network connection, to avoid confusion about terms
// it can be either a full peer and or a client (or sth else)
type Vertex struct {
	wsConn   *websocket.Conn
	address  string
	vertexid uuid.UUID
	name     string
	//channels
	in_read   chan (Msg)
	out_write chan (Msg)
	handshake bool
	isPeer    bool
	isClient  bool
	//currently expected to validate
	//isLeader  bool
}

// func broadcast(textmsg string) {
// 	for _, cl := range vertexs {
// 		log.Println("send to ", cl.vertexid, textmsg)
// 		xmsg := Msg{Type: "chat", Value: textmsg}
// 		cl.out_write <- xmsg
// 	}
// }

func readLoop(vertex Vertex) {
	for {
		// contiously read in a message and put on channel
		_, p, err := vertex.wsConn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("msg received: ", string(p)+" "+vertex.name)
		msg := Msg{}
		err = json.Unmarshal([]byte(string(p)), &msg)
		if err == nil {
			fmt.Printf("type: %s\n", msg.Type)
			fmt.Printf("value: %s\n", msg.Value)
		}
		msg.Sender = vertex.vertexid
		msg.Time = time.Now()
		log.Println("put msg ", msg)

		vertex.in_read <- msg
	}
}

func writeLoop(vertex Vertex) {
	//log.Println("writeLoop ", vertex)
	for {
		//log.Println("writeLoop...")
		select {
		case msgOut := <-vertex.out_write:
			fmt.Println(msgOut)
			log.Println("msgout  ", msgOut)
			//xmsg := Msg{Type: "name", Value: msgOut.Value + "|registered"}
			msgByte, _ := json.Marshal(msgOut)
			err := vertex.wsConn.WriteMessage(1, msgByte)
			if err != nil {
				log.Println("error writing to ", vertex.vertexid)
			}
		case <-time.After(time.Second * 50):
			fmt.Println("TIMEOUT: out_write")
		}

	}
}
