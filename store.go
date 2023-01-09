package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/horizonledger/protocol"
)

func writeState(state *State) {
	log.Println("write state. messages: ", len(state.MsgHistory))
	log.Println("write state. messages: ", state.MsgHistory[len(state.MsgHistory)-1:])
	jsonData, _ := json.MarshalIndent(state, "", " ")
	//fmt.Println(string(jsonData))
	_ = ioutil.WriteFile(stateFile, jsonData, 0644)
}

func initStorage() State {
	fmt.Println("init storage")
	var emptyHistory []protocol.Msg
	state := State{LastUpdate: time.Now(), MsgHistory: emptyHistory}
	writeState(&state)
	return state
}

func storageInited() bool {
	if _, err := os.Stat(stateFile); err == nil {
		return true
	} else {
		return false
	}
}

func loadStorage() State {

	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}
	state := State{}
	if err := json.Unmarshal(data, &state); err != nil {
		panic(err)
	}
	return state
}

func saveState(state *State) {
	//TODO store only if state has changed
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			writeState(state)
		case <-quit:
			ticker.Stop()
			return
		}
	}
}
