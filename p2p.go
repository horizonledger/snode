package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func broadcast(textmsg string) {
	for _, cl := range clients {
		log.Println("send to ", cl.clientid, textmsg)
		xmsg := Msg{Type: "chat", Value: textmsg}
		cl.out_write <- xmsg
	}
}

func readLoop(client Client) {
	for {
		// contiously read in a message and put on channel
		_, p, err := client.wsConn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("msg received: ", string(p))
		msg := Msg{}
		err = json.Unmarshal([]byte(string(p)), &msg)
		if err == nil {
			fmt.Printf("type: %s\n", msg.Type)
			fmt.Printf("value: %s\n", msg.Value)
		}
		msg.Sender = client.clientid
		msg.Time = time.Now()
		log.Println("put msg ", msg)

		client.in_read <- msg
	}
}

func writeLoop(client Client) {
	log.Println("writeLoop ")
	for {
		log.Println("writeLoop2 ")
		msgOut := <-client.out_write
		log.Println("msgout  ", msgOut)
		//xmsg := Msg{Type: "name", Value: msgOut.Value + "|registered"}
		msgByte, _ := json.Marshal(msgOut)
		client.wsConn.WriteMessage(1, msgByte)
	}
}
