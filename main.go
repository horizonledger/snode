package main

import (
	"encoding/json"
	"flag"
	"fmt"
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
	//port for websocket used by clients and peers
	Port int
	//SlotID int
	// CreateGenesis bool
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func getConfig(conffile string) Config {

	log.Println("config file ", conffile)

	if _, err := os.Stat(conffile); os.IsNotExist(err) {
		log.Println("config file does not exist. create a file named ", conffile)
		//return nil
	}

	content, err := ioutil.ReadFile(conffile)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	fmt.Println("... ", string(content))

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	fmt.Println(">> ", config.Port)

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

}

func main() {
	portArg := flag.Int("port", 0, "port number")
	configFileArg := flag.String("config", "", "config file")
	flag.Parse()
	//println("portArg ", *portArg)
	//println("configFileArg ", *configFileArg)

	cfile := "config.json"
	if *configFileArg != "" {
		cfile = *configFileArg
	}
	config := getConfig(cfile)

	//override with flag
	if *portArg != 0 {
		config.Port = *portArg
	}

	log.Println(config)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	startupNode(config)
	<-quit // This will block until you manually exists with CRl-C
	log.Println("\nnode exiting")
}
