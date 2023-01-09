package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/horizonledger/protocol"
)

// Vertex is a wrapper around a network connection, to avoid confusion about terms
// it can be either a full peer and or a client (or sth else)
type Vertex struct {
	wsConn   *websocket.Conn
	address  string
	vertexid uuid.UUID
	name     string
	//channels
	in_read   chan (protocol.Msg)
	out_write chan (protocol.Msg)
	handshake bool
	isPeer    bool
	isClient  bool
	//currently expected to validate
	//isLeader  bool
}

func broadcast(state *State, textmsg string) {
	for _, cl := range state.vertexs {
		log.Println("send to ", cl.vertexid, textmsg)
		xmsg := protocol.Msg{Type: "chat", Value: textmsg}
		cl.out_write <- xmsg
	}
}

func readLoop(vertex *Vertex) {
	for {
		// contiously read in a message and put on channel
		_, p, err := vertex.wsConn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("bytes received: ", string(p)+" "+vertex.name)
		msg := protocol.ParseMessageFromBytes(p)
		log.Println("msg received: ", msg.Type+" "+vertex.name)
		msg.Sender = vertex.vertexid
		msg.Time = time.Now()
		log.Println("put msg in chan: ", msg)

		vertex.in_read <- msg
	}
}

func writeLoop(vertex *Vertex) {
	//log.Println("writeLoop ", vertex)
	for {
		//log.Println("writeLoop...")
		select {
		case msgOut := <-(*vertex).out_write:
			fmt.Println(msgOut)
			log.Println("msgout  ", msgOut)

			msgBytes := protocol.ParseMessageToBytes(msgOut)
			err := vertex.wsConn.WriteMessage(1, msgBytes)
			if err != nil {
				log.Println("error writing to ", vertex.vertexid)
			}
		case <-time.After(time.Second * 50):
			fmt.Println("TIMEOUT: nothing to write on loop")
		}

	}
}
