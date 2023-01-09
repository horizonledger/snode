package main

import (
	"strconv"

	"github.com/horizonledger/protocol"
	log "github.com/sirupsen/logrus"
)

// TODO separate different modalities
// requests
// transactions
const STATUS = "STATUS"
const HNDPEER = "HNDPEER"
const HNDCLIENT = "HNDCLIENT"
const STATE = "state"
const REQHEIGHT = "REQHEIGHT"
const REPHEIGHT = "REPHEIGHT"
const CHAT = "chat"
const NAME = "name"

func handleMsg(state *NodeState, vertex *Vertex, msg protocol.Msg) {

	log.Debug("handle msg ", msg)
	log.Debug("type >> ", msg.Type)

	switch msg.Type {
	case STATUS:
		log.Debug("STATUS received ")
		log.Debug(">> ", msg.Value)
		//who is leader and follower?
		//if follower and new state then
		//pushState(vertex.wsConn)

	case HNDPEER:
		//TODO check not already connected
		//TODO pubkey exchange here
		if vertex.handshake {
			log.Info("handle handshake already")
		} else {
			log.Info("handle handshake")
			xmsg := protocol.Msg{Type: "HNDSHAKEPEER", Value: "confirm"}
			vertex.out_write <- xmsg
			vertex.handshake = true
			vertex.isPeer = true
			vertex.isClient = false
		}

	case HNDCLIENT:
		if vertex.handshake {
			log.Debug("handle handshake already")
		} else {
			log.Debug("handle handshake")
			xmsg := protocol.Msg{Type: "HNDSHAKECLIENT", Value: "confirm"}
			vertex.out_write <- xmsg
			vertex.handshake = true
			vertex.isPeer = false
			vertex.isClient = true

			//separate message. client needs to pull
			pushState(*vertex)

			//push uuid to client
			infoMsg := protocol.Msg{Type: "uuid", Value: vertex.vertexid.String()}
			vertex.out_write <- infoMsg

		}

	case STATE:
		//TODO from index i to index j
		log.Debug("handle state")
		pushState(*vertex)

	case REQHEIGHT:
		n := len(nodestate.msgstate.MsgHistory)
		xmsg := protocol.Msg{Type: REPHEIGHT, Value: strconv.Itoa(n)}
		vertex.out_write <- xmsg

	case CHAT:
		log.Debug("handle chat")

		cid := vertex.vertexid.String()
		log.Debug("vertex name: ", vertex.name)
		log.Debug("vertexid: ", cid)
		if vertex.name != "default" {
			cid = vertex.name
		}
		textmsg := cid + ": " + msg.Value
		log.Debug("textmsg ", textmsg)
		//broadcast
		log.Debug("vertexs len %v", len(state.vertexs))
		//TODO fix
		broadcast(state, textmsg)

		//testing
		// xmsg := protocol.Msg{Type: "chat", Value: textmsg}
		// vertex.out_write <- xmsg

	//register name only, no transfers yet
	case NAME:
		//TODO check duplicate names on registration
		log.Debug("handle name")
		log.Info("set name to: " + msg.Value)
		//TODO check if already registered
		//TODO name change doesnt persist, need to share reference to vertex map
		vertex.name = msg.Value
		state.vertexs[vertex.vertexid] = *vertex
		//state.
		log.Info("name now : " + vertex.name + " " + vertex.vertexid.String())

		xmsg := protocol.Msg{Type: "name", Value: msg.Value + "|registered"}
		vertex.out_write <- xmsg
		// msgByte, _ := json.Marshal(xmsg)
		// //cl.wsConn.WriteMessage(messageType, msgByte)
		// vertex.wsConn.WriteMessage(1, msgByte)
	default:
		return
	}
	//save state
	log.Debug("save state")
	nodestate.msgstate.MsgHistory = append(nodestate.msgstate.MsgHistory, msg)

}
