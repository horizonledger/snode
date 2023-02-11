package main

import (
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/horizonledger/protocol"
	. "github.com/horizonledger/protocol"
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
	in_read   chan (Gen)
	out_write chan (Gen)
	handshake bool
	isPeer    bool
	isClient  bool
	//TODO will be blockheight later
	height time.Time
	//currently expected to validate
	//isLeader  bool
}

// func broadcast(state *NodeState, textmsg string) {
// 	for _, cl := range state.vertexs {
// 		log.Debug("send to ", cl.vertexid, textmsg)
// 		//TODO in separate generic function which translates Msg to Gen
// 		xmsg := Msg{Type: "chat", Value: textmsg}
// 		jxmsg := ParseMessageToBytes(xmsg)
// 		gen := Gen{Type: "Msg", Value: jxmsg}
// 		cl.out_write <- gen
// 	}
// }

func readLoop(vertex *Vertex) {
	for {
		// contiously read in a message and put on channel
		_, p, err := vertex.wsConn.ReadMessage()
		if err != nil {
			log.Warn("error read loop ", err)
			return
		}
		log.Debug("bytes received: ", string(p)+" "+vertex.name)
		//msg := ParseMessageFromBytes(p)
		genmsg := ParseGenFromBytes(p)
		log.Debug("gen received: ", genmsg.Type+" "+vertex.vertexid.String())
		//TODO
		if vertex.name != "default" {
			genmsg.Sender = vertex.name
		}
		//vertex.vertexid
		genmsg.Time = time.Now()

		log.Debug("msg in chan: ", genmsg)

		//gengmsg.Value = x
		//z := ParseMessageFromBytes(genmsg.Value)
		vertex.in_read <- genmsg
	}
}

func writeLoop(nodestate *NodeState, vertex *Vertex) {
	//log.Println("writeLoop ", vertex)
	for {
		//log.Println("writeLoop...")
		select {
		case msgOut := <-(*vertex).out_write:
			log.Trace("msg out  ", msgOut)
			//TODO handle transactions

			//msgBytes := ParseMessageToBytes(msgOut)
			genmsgBytes := ParseGenToBytes(msgOut)
			err := vertex.wsConn.WriteMessage(1, genmsgBytes)
			if err != nil {
				log.Error("error writing to ", vertex.vertexid)
				log.Error("close connection ", vertex.vertexid)
				vertex.wsConn.Close()
				delete(nodestate.vertexs, vertex.vertexid)

			}
		case <-time.After(time.Second * 50):
			log.Debug("TIMEOUT: nothing to write on loop")
		}

	}
}

func handleSub(nodestate *NodeState, vertex *Vertex, msg protocol.Msg) {
	//TODO which topic is being subscribed to?
	log.Info("handle sub")

	topic := msg.Value
	log.Info("handle sub ", topic)

	//TODO validate topic is valid
	hookVertexToSub(vertex, topic)

}

func readHandler(nodestate *NodeState, vertex *Vertex) {
	for {
		genmsg := <-(*vertex).in_read
		log.Info("readHandler ", genmsg)
		log.Info("type ", genmsg.Type)
		if genmsg.Type == "Msg" {
			msg := ParseMessageFromBytes(genmsg.Value)
			msg.Time = time.Now()
			//msg.Sender = vertex.vertexid
			msg.Sender = vertex.name
			log.Info("msg handler ", msg)
			//TODO
			// switch msg.Category {
			// case "REQ":
			// 	log.Info("request...")
			// 	handleRequest(nodestate, vertex, msg)
			// case "REP":
			// 	log.Info("reply...")
			// case "PUB":
			// 	log.Info("pub...")
			// case "SUB":
			// 	log.Info("sub...")
			// 	handleSub(nodestate, vertex, msg)
			// default:
			// 	log.Info("unknown category")

			// }
			//TODO
			//handleMsg(nodestate, vertex, msg)

		} else if genmsg.Type == "NameTx" {
			log.Info("handle tx..")
			tx := ParseTxFromBytes(genmsg.Value)
			log.Info("> tx..", tx)
			handleTx(nodestate, vertex, tx)
			//TODO time
		}

	}
}

type Pubsub struct {
	mu sync.RWMutex
	//topic is by string
	subs map[string][]chan Gen
	//closed bool
}

func (ps *Pubsub) Subscribe(topic string, ch chan Gen) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.subs[topic] = append(ps.subs[topic], ch)
}

func (ps *Pubsub) UnSubscribe(topic string, ch chan Gen) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	delete(ps.subs, topic)
}

func NewPubsub() *Pubsub {
	ps := &Pubsub{}
	ps.subs = make(map[string][]chan Gen)
	return ps
}

func (ps *Pubsub) Publish(topic string, msg Gen) {
	log.Println("publish...")
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// if ps.closed {
	// 	return
	//   }

	for _, ch := range ps.subs[topic] {
		log.Println("publish... ", ch, topic, msg)
		ch <- msg
	}
}

//not used
// func (ps *Pubsub) Close() {
// 	ps.mu.Lock()
// 	defer ps.mu.Unlock()

// 	if !ps.closed {
// 	  ps.closed = true
// 	  for _, subs := range ps.subs {
// 		for _, ch := range subs {
// 		  close(ch)
// 		}
// 	  }
// 	}
//   }
