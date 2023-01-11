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
	in_read   chan (protocol.Gen)
	out_write chan (protocol.Gen)
	handshake bool
	isPeer    bool
	isClient  bool
	//TODO will be blockheight later
	height time.Time
	//currently expected to validate
	//isLeader  bool
}

func broadcast(state *NodeState, textmsg string) {
	for _, cl := range state.vertexs {
		log.Debug("send to ", cl.vertexid, textmsg)
		//TODO in separate generic function which translates Msg to Gen
		xmsg := protocol.Msg{Type: "chat", Value: textmsg}
		jxmsg := protocol.ParseMessageToBytes(xmsg)
		gen := protocol.Gen{Type: "Msg", Value: jxmsg}
		cl.out_write <- gen
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
		//msg := protocol.ParseMessageFromBytes(p)
		genmsg := protocol.ParseGenFromBytes(p)
		log.Debug("gen received: ", genmsg.Type+" "+vertex.vertexid.String())
		genmsg.Sender = vertex.vertexid
		genmsg.Time = time.Now()
		log.Debug("put msg in chan: ", genmsg)

		//gengmsg.Value = x
		//z := protocol.ParseMessageFromBytes(genmsg.Value)
		vertex.in_read <- genmsg
	}
}

func writeLoop(vertex *Vertex) {
	//log.Println("writeLoop ", vertex)
	for {
		//log.Println("writeLoop...")
		select {
		case msgOut := <-(*vertex).out_write:
			log.Debug("msgout  ", msgOut)
			//TODO handle transactions

			//msgBytes := protocol.ParseMessageToBytes(msgOut)
			genmsgBytes := protocol.ParseGenToBytes(msgOut)
			err := vertex.wsConn.WriteMessage(1, genmsgBytes)
			if err != nil {
				log.Error("error writing to ", vertex.vertexid)
			}
		case <-time.After(time.Second * 50):
			log.Debug("TIMEOUT: nothing to write on loop")
		}

	}
}

func readHandler(nodestate *NodeState, vertex *Vertex) {
	for {
		genmsg := <-(*vertex).in_read
		log.Debug("readHandler.. ", genmsg)
		if genmsg.Type == "Msg" {
			msg := protocol.ParseMessageFromBytes(genmsg.Value)
			log.Debug("msg.. ", msg)
			handleMsg(nodestate, vertex, msg)
		} else if genmsg.Type == "NameTx" {
			log.Info("handle tx..")
			tx := protocol.ParseTxFromBytes(genmsg.Value)
			log.Info("> tx..", tx)
			handleTx(nodestate, vertex, tx)
		}

	}
}
