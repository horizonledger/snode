package main

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/horizonledger/protocol"
	log "github.com/sirupsen/logrus"
)

// TODO separate different modalities
// requests
// transactions
// system level messages
const STATUS = "STATUS"
const HNDPEER = "HNDPEER"
const HNDCLIENT = "HNDCLIENT"
const STATE = "state"
const REQSTATE = "REQSTATE"
const REPSTATE = "REPSTATE"
const REQHEIGHT = "REQHEIGHT"
const REPHEIGHT = "REPHEIGHT"

// app level messages
const CHAT = "chat"
const NAME = "name"

func containsName(m map[string]uuid.UUID, v string) bool {
	_, ok := m[v]
	return ok

	// for _, x := range m {
	// 	if x == v {
	// 		return true
	// 	}
	// }
	// return false
}

func handleTx(state *NodeState, vertex *Vertex, tx protocol.NameTx) {

	switch tx.Type {
	case NAME:
		log.Debug("handletx .. name")
		log.Debug(tx.Action)
		log.Debug(tx.Name)
	}

}

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
			vertex.out_write <- protocol.MsgToGen(xmsg)
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
			vertex.out_write <- protocol.MsgToGen(xmsg)
			vertex.handshake = true
			vertex.isPeer = false
			vertex.isClient = true

			//separate message. client needs to pull
			pushState(*vertex)

			//push uuid to client
			infoMsg := protocol.Msg{Type: "uuid", Value: vertex.vertexid.String()}
			vertex.out_write <- protocol.MsgToGen(infoMsg)

		}

	case REQSTATE:
		//TODO push only height i to j
		log.Info("state request...")
		pushState(*vertex)

	case REPSTATE:
		//TODO merge to local state
		log.Info("repstate... todo merge in")
		//log.Info("msg.Value ", msg.Value)
		msghist := msgHistFromJson([]byte(msg.Value))
		//log.Info("hist ", msghist)
		log.Info("len ", len(msghist))
		log.Info("last msg ", msghist[len(msghist)-1])
		log.Info("last msg time: ", msghist[len(msghist)-1].Time)

		c := 0
		if len(nodestate.msgstate.MsgHistory) == 0 {
			nodestate.msgstate.MsgHistory = msghist
			c = len(msghist)
		} else {
			ourLast := lastMsgTime(nodestate.msgstate.MsgHistory)

			for _, hmsg := range msghist {
				//TODO check time
				timeDiff := hmsg.Time.Sub(ourLast)
				log.Info("timeDiff ", timeDiff)
				if timeDiff.Seconds() > 0 {
					nodestate.msgstate.MsgHistory = append(nodestate.msgstate.MsgHistory, hmsg)
					c++
				}
			}
			log.Debug("appended ", c)
		}

		log.Info("history now ", len(nodestate.msgstate.MsgHistory))
		log.Info("appened ", c)

	case REQHEIGHT:
		//n := len(nodestate.msgstate.MsgHistory)
		n := len(msgstate.MsgHistory)
		if n > 0 {
			lastMsgTime := msgstate.MsgHistory[n-1].Time
			xmsg := protocol.Msg{Type: REPHEIGHT, Value: lastMsgTime.String()}
			vertex.out_write <- protocol.MsgToGen(xmsg)
		}

	case REPHEIGHT:
		//TODO in separate routine
		//TODO request the delta i to j
		log.Info("the height of the peer is ", msg.Value)

		syncHistory(vertex)

		//log.Info("our height is ", len(nodestate.msgstate.MsgHistory))
		//TODO cant use message length we time of last message
		// h, _ := strconv.Atoi(msg.Value)
		// if len(nodestate.msgstate.MsgHistory) < h {
		// 	log.Info("request state, since we're behind")
		// 	xmsg := protocol.Msg{Type: REQSTATE, Value: strconv.Itoa(0)}
		// 	vertex.out_write <- xmsg
		// }

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
		log.Debug("vertexs len ", len(state.vertexs))
		//TODO fix
		broadcast(state, textmsg)

		//testing
		// xmsg := protocol.Msg{Type: "chat", Value: textmsg}
		// vertex.out_write <- xmsg

	//register name only, no transfers yet
	case NAME:
		//TODO check duplicate names on registration
		newName := msg.Value
		if containsName(nodestate.unames, newName) {
			log.Debug("name exists")
			//TODO better message here
			xmsg := protocol.Msg{Type: NAME, Value: newName + "|exists already"}
			vertex.out_write <- protocol.MsgToGen(xmsg)
		} else {

			log.Debug("handle name")
			log.Info("set name to: " + msg.Value)
			//TODO check if already registered
			//TODO name change doesnt persist, need to share reference to vertex map
			vertex.name = msg.Value
			state.vertexs[vertex.vertexid] = *vertex
			//state.
			log.Info("name now : " + vertex.name + " " + vertex.vertexid.String())
			//set owner
			//TODO make this pubkey
			nodestate.unames[newName] = vertex.vertexid

			xmsg := protocol.Msg{Type: NAME, Value: msg.Value + "|registered"}
			vertex.out_write <- protocol.MsgToGen(xmsg)

		}
		// msgByte, _ := json.Marshal(xmsg)
		// //cl.wsConn.WriteMessage(messageType, msgByte)
		// vertex.wsConn.WriteMessage(1, msgByte)
	default:
		return
	}

	//save state
	if msg.Type == CHAT || msg.Type == NAME {
		log.Debug("save state")
		nodestate.msgstate.MsgHistory = append(nodestate.msgstate.MsgHistory, msg)
	}

	//we could save system type messages as well
	//storing state messages introduces recursion never store that message

}

func msgHistToJson(msghist []protocol.Msg) []byte {
	msgData, _ := json.MarshalIndent(msghist, "", " ")
	return msgData
}

func msgHistFromJson(p []byte) []protocol.Msg {
	var msghist []protocol.Msg
	err := json.Unmarshal(p, &msghist)
	if err != nil {
		log.Println("couldnt parse message")
	}

	return msghist
}

func pushState(vertex Vertex) {
	msgData := msgHistToJson(nodestate.msgstate.MsgHistory)
	statemsg := protocol.Msg{Type: REPSTATE, Value: string(msgData)}
	vertex.out_write <- protocol.MsgToGen(statemsg)
	//stateData, _ := json.MarshalIndent(statemsg, "", " ")
	//_ = vertex.wsConn.WriteMessage(1, []byte(stateData))
}

func lastMsgTime(msghist []protocol.Msg) time.Time {
	n := len(msghist)
	lastEl := msghist[n-1]
	return lastEl.Time
}
