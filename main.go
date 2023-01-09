package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"

	"net/http"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

type Config struct {
	NodeAlias string
	Verbose   bool
	//port for websocket used by clients and peers
	Port       int
	NodeID     int
	InitVertex []string
	StateFile  string
	//SlotID int
	// CreateGenesis bool
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func getConfig(conffile string) Config {

	log.Info("config file ", conffile)

	if _, err := os.Stat(conffile); os.IsNotExist(err) {
		log.Println("config file does not exist. create a file named ", conffile)
		//return nil
	}

	content, err := ioutil.ReadFile(conffile)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	log.Debug("... ", string(content))

	var config Config
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	log.Debug(">> ", config.Port)

	return config

}

func setupRoutes() {
	// fs := http.FileServer(http.Dir("./static"))
	// http.Handle("/", fs)

	http.HandleFunc("/ws", serveWs)
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

	// 	log.Trace("Something very low level.")
	// log.Debug("Useful debugging information.")
	// log.Info("Something noteworthy happened!")
	// log.Warn("You should probably take a look at this.")
	// log.Error("Something failed but I'm not quitting.")

	log.Info(config)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	startupNode(config)
	<-quit // This will block until you manually exists with CRl-C
	log.Warn("\nnode exiting")
}
