//TODO chat history
// integrate app
// wallet
// online status
// profiile

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

type Msg struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Client struct {
	wsConn   *websocket.Conn
	clientid uuid.UUID
	name     string
	//channels
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var clients = map[uuid.UUID]Client{}

func broadcast(textmsg string) {
	for _, cl := range clients {
		fmt.Printf("send to %v\n", cl.clientid)
		xmsg := Msg{Type: "chat", Value: textmsg}
		msgByte, _ := json.Marshal(xmsg)
		cl.wsConn.WriteMessage(1, msgByte)
	}
}

func reader(client Client) {
	for {
		// read in a message
		_, p, err := client.wsConn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("msg: ", string(p))

		msg := Msg{}
		err = json.Unmarshal([]byte(string(p)), &msg)
		if err == nil {
			fmt.Printf("type: %s\n", msg.Type)
			fmt.Printf("value: %s\n", msg.Value)
		}

		if msg.Type == "chat" {
			cid := client.clientid.String()
			fmt.Printf("client.name %v", client.name)
			fmt.Printf("cid %v", cid)
			if client.name != "default" {
				cid = client.name
			}
			textmsg := cid + ": " + msg.Value
			//broadcast
			fmt.Printf("clients len %v", len(clients))
			broadcast(textmsg)

		} else if msg.Type == "name" {
			fmt.Print("name " + msg.Value)
			//TOOD check if already registered
			client.name = msg.Value

			xmsg := Msg{Type: "name", Value: "name registered"}
			msgByte, _ := json.Marshal(xmsg)
			//cl.wsConn.WriteMessage(messageType, msgByte)
			client.wsConn.WriteMessage(1, msgByte)
		}

	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client Connected")
	var newuid = uuid.Must(uuid.NewV4())
	client := Client{wsConn: ws, clientid: newuid, name: "default"}
	clients[newuid] = client
	log.Println("uid ", newuid)

	for _, cl := range clients {
		welcome_msg := Msg{Type: "chat", Value: newuid.String() + " entered"}
		msgByte, _ := json.Marshal(welcome_msg)
		fmt.Printf("send to %v %v\n", cl.clientid, string(msgByte))
		cl.wsConn.WriteMessage(1, msgByte)
	}

	_ = ws.WriteMessage(1, []byte("handshake from server"))

	infoMsg := Msg{Type: "uuid", Value: newuid.String()}
	replyByte, _ := json.Marshal(infoMsg)

	_ = ws.WriteMessage(1, replyByte)

	reader(client)
}

func setupRoutes() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	fmt.Println("running on 8080")
	clients = make(map[uuid.UUID]Client)
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))

}
