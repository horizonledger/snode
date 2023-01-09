package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/horizonledger/protocol"
)

func writeState(stateFile string, state *MsgState) {
	log.Println("write state. messages: ", len(state.MsgHistory))
	//log.Println("write state. last message: ", state.MsgHistory[len(state.MsgHistory)-1:])
	jsonData, _ := json.MarshalIndent(state, "", " ")
	//log.Debug(string(jsonData))
	_ = ioutil.WriteFile(stateFile, jsonData, 0644)
}

func initStorage(stateFile string) MsgState {
	log.Debug("init storage")
	var emptyHistory []protocol.Msg
	state := MsgState{LastUpdate: time.Now(), MsgHistory: emptyHistory}
	writeState(stateFile, &state)
	return state
}

func storageInited(stateFile string) bool {
	if _, err := os.Stat(stateFile); err == nil {
		return true
	} else {
		return false
	}
}

func loadStorage(stateFile string) MsgState {

	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}
	state := MsgState{}
	if err := json.Unmarshal(data, &state); err != nil {
		panic(err)
	}
	return state
}

func saveState(stateFile string, state *MsgState) {
	//TODO store only if state has changed
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			writeState(stateFile, state)
		case <-quit:
			ticker.Stop()
			return
		}
	}
}
