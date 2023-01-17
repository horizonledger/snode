package main

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/horizonledger/protocol"
	"github.com/horizonledger/protocol/crypto"
	log "github.com/sirupsen/logrus"
)

// TODO separate different modalities
// requests
// transactions
// system level messages
const STATUS = "STATUS"
const PING = "PING"
const PONG = "PONG"
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
const INFO = "info"

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

func handleNameRegister(nodestate *NodeState, vertex *Vertex, newName string) {
	//register name only, no transfers yet

	//check duplicate names on registration
	if containsName(nodestate.unames, newName) {
		log.Debug("name exists")
		//TODO better message here
		xmsg := protocol.Msg{Type: NAME, Value: newName + "|exists already"}
		vertex.out_write <- protocol.MsgToGen(xmsg)
	} else {

		log.Debug("handle name")
		log.Info("set name to: " + newName)
		//TODO check if already registered
		//TODO name change doesnt persist, need to share reference to vertex map
		vertex.name = newName
		nodestate.vertexs[vertex.vertexid] = *vertex
		//state.
		log.Info("name now : " + vertex.name + " " + vertex.vertexid.String())
		//set owner
		//TODO make this pubkey
		nodestate.unames[newName] = vertex.vertexid
		//TODO txid
		xmsg := protocol.Msg{Type: NAME, Value: newName + "|registered"}
		vertex.out_write <- protocol.MsgToGen(xmsg)

	}
}

func handleTx(state *NodeState, vertex *Vertex, tx protocol.NameTx) {

	switch tx.Type {
	case NAME:
		//TODO check action
		log.Debug("handletx .. name")
		// log.Debug(tx.Action)
		// log.Debug(tx.Name)
		// log.Debug(tx.Signature)
		// log.Debug(tx.SenderPubkey)
		// log.Debug(tx)
		pub := crypto.PubKeyFromHex(tx.SenderPubkey)
		sigValid := crypto.VerifySignedTx(pub, tx)
		log.Println("verified tx ", sigValid)
		if sigValid {
			handleNameRegister(state, vertex, tx.Name)
		}
	}

}

func handleRequest(state *NodeState, vertex *Vertex, msg protocol.Msg) {

	switch msg.Type {

	case PING:
		xmsg := protocol.Msg{Type: PONG, Value: "confirm"}
		vertex.out_write <- protocol.MsgToGen(xmsg)

	case HNDCLIENT:
		if vertex.handshake {
			log.Debug("handled handshake already")
		} else {
			log.Debug("handle handshake")
			xmsg := protocol.Msg{Type: "HNDSHAKECLIENT", Value: "confirm"}
			vertex.out_write <- protocol.MsgToGen(xmsg)
			vertex.handshake = true
			vertex.isPeer = false
			vertex.isClient = true

			//separate message. client needs to pull
			//pushState(*vertex)

			//push uuid to client
			infoMsg := protocol.Msg{Type: "uuid", Value: vertex.vertexid.String()}
			vertex.out_write <- protocol.MsgToGen(infoMsg)

		}

	case CHAT:
		log.Info("handle chat")

		//TODO add author name
		xmsg := protocol.Msg{Category: "PUB", Type: "CHAT", Value: msg.Value}
		nodestate.pubsub.Publish("chat", protocol.MsgToGen(xmsg))

		cid := vertex.vertexid.String()
		log.Debug("vertex name: ", vertex.name)
		log.Debug("vertexid: ", cid)
		if vertex.name == "default" {
			log.Info("need to register")
			//cid = vertex.name
			xmsg := protocol.Msg{Type: INFO, Value: "register name first"}
			vertex.out_write <- protocol.MsgToGen(xmsg)

		} else {
			textmsg := vertex.name + ": " + msg.Value
			log.Debug("textmsg ", textmsg)
			//broadcast
			log.Debug("vertexs len ", len(state.vertexs))

			//TODO fix publish here
			//broadcast(state, textmsg)
		}

	default:
		log.Warn("unknwon request type")
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
			xmsg := protocol.Msg{Type: HNDPEER, Value: "confirm"}
			vertex.out_write <- protocol.MsgToGen(xmsg)
			vertex.handshake = true
			vertex.isPeer = true
			vertex.isClient = false
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

		//testing
		// xmsg := protocol.Msg{Type: "chat", Value: textmsg}
		// vertex.out_write <- xmsg

	default:
		return
	}

	//save state
	//TODO this will change with tx
	//store chat message in separate state not synchronized
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
		log.Error("couldnt parse message ", string(p))
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
