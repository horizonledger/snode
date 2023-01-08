package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

type Config struct {
	NodeAlias string
	Verbose   bool
	NodePort  int
	WebPort   int
	SlotID    int
	// CreateGenesis bool
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func readHandler(vertex Vertex) {
	for {
		msg := <-vertex.in_read
		handleMsg(vertex, msg)
	}
}

func getConfig() Config {

	conffile := "config.json"
	log.Println("config file ", conffile)

	if _, err := os.Stat(conffile); os.IsNotExist(err) {
		log.Println("config file does not exist. create a file named ", conffile)
		//return nil
	}

	content, err := ioutil.ReadFile(conffile)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	return config

}

func setupRoutes() {
	// fs := http.FileServer(http.Dir("./static"))
	// http.Handle("/", fs)

	http.HandleFunc("/ws", serveWs)
}

var (
	//env  *string
	port       *int
	configFile *string
)

func init() {
	//env = flag.String("env", "development", "current environment")
	port = flag.Int("port", 8000, "port number")
	configFile = flag.String("config", "", "port number")
}

func main() {
	flag.Parse()
	log.Println("?? ", *port)
	//config := getConfig()
	config := Config{WebPort: *port}
	log.Println(config)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	startupNode(config)
	<-quit // This will block until you manually exists with CRl-C
	log.Println("\nnode exiting")
}
