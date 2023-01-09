package main

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/horizonledger/protocol"
	log "github.com/sirupsen/logrus"
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

func broadcast(state *NodeState, textmsg string) {
	for _, cl := range state.vertexs {
		log.Debug("send to ", cl.vertexid, textmsg)
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
		log.Debug("bytes received: ", string(p)+" "+vertex.name)
		msg := protocol.ParseMessageFromBytes(p)
		log.Debug("msg received: ", msg.Type+" "+vertex.vertexid.String())
		msg.Sender = vertex.vertexid
		msg.Time = time.Now()
		log.Debug("put msg in chan: ", msg)

		vertex.in_read <- msg
	}
}

func writeLoop(vertex *Vertex) {
	//log.Println("writeLoop ", vertex)
	for {
		//log.Println("writeLoop...")
		select {
		case msgOut := <-(*vertex).out_write:
			log.Debug("msgout  ", msgOut)

			msgBytes := protocol.ParseMessageToBytes(msgOut)
			err := vertex.wsConn.WriteMessage(1, msgBytes)
			if err != nil {
				log.Error("error writing to ", vertex.vertexid)
			}
		case <-time.After(time.Second * 50):
			log.Debug("TIMEOUT: nothing to write on loop")
		}

	}
}
