package main

import (
	"fmt"
	"log"
)

func handleMsg(state State, vertex Vertex, msg Msg) {

	log.Println("handle msg ", msg)
	log.Println("type >> ", msg.Type)

	switch msg.Type {
	case "STATUS":
		log.Println("STATUS received ")
		log.Println(">> ", msg.Value)
		//who is leader and follower?
		//if follower and new state then
		//pushState(vertex.wsConn)

	case "HNDPEER":
		//TODO check not already connected
		//TODO pubkey exchange here
		if vertex.handshake {
			log.Println("handle handshake already")
		} else {
			log.Println("handle handshake")
			xmsg := Msg{Type: "HNDSHAKEPEER", Value: "confirm"}
			vertex.out_write <- xmsg
			vertex.handshake = true
			vertex.isPeer = true
			vertex.isClient = false
		}

	case "HNDCLIENT":
		if vertex.handshake {
			log.Println("handle handshake already")
		} else {
			log.Println("handle handshake")
			xmsg := Msg{Type: "HNDSHAKECLIENT", Value: "confirm"}
			vertex.out_write <- xmsg
			vertex.handshake = true
			vertex.isPeer = false
			vertex.isClient = true

			pushState(vertex)

			//push uuid to client
			infoMsg := Msg{Type: "uuid", Value: vertex.vertexid.String()}
			vertex.out_write <- infoMsg

		}

	case "state":
		log.Println("handle state")

	case "chat":
		log.Println("handle chat")

		cid := vertex.vertexid.String()
		log.Println("vertex name: ", vertex.name)
		log.Println("vertexid: ", cid)
		if vertex.name != "default" {
			cid = vertex.name
		}
		textmsg := cid + ": " + msg.Value
		log.Println("textmsg ", textmsg)
		//broadcast
		fmt.Printf("vertexs len %v", len(state.vertexs))
		//TODO fix
		//broadcast(textmsg)

		//testing
		xmsg := Msg{Type: "chat", Value: textmsg}
		vertex.out_write <- xmsg

	case "name":
		//TODO check duplicate names on registration
		log.Println("handle name")
		fmt.Println("set name to: " + msg.Value)
		//TODO check if already registered
		//TODO name change doesnt persist
		vertex.name = msg.Value
		state.vertexs[vertex.vertexid] = vertex
		//state.
		fmt.Println("name now : " + vertex.name)

		xmsg := Msg{Type: "name", Value: msg.Value + "|registered"}
		vertex.out_write <- xmsg
		// msgByte, _ := json.Marshal(xmsg)
		// //cl.wsConn.WriteMessage(messageType, msgByte)
		// vertex.wsConn.WriteMessage(1, msgByte)
	default:
		return
	}
	//save state
	state.MsgHistory = append(state.MsgHistory, msg)

}
