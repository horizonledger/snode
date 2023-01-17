package main

import (
	"testing"

	"github.com/gofrs/uuid"
)

func TestBasicTransfer(t *testing.T) {

	//TODO setup node
	nodestate = NodeState{isLeader: true, msgstate: msgstate, vertexs: make(map[uuid.UUID]Vertex)}
	nodestate.cashbalance = make(map[string]uint64)

	nodestate.cashbalance["a"] = 200
	nodestate.cashbalance["b"] = 100

	if nodestate.cashbalance["b"] != 100 {
		t.Error("wrong balance")
	}

	transferCash(nodestate, "a", "b", 50)

	if nodestate.cashbalance["b"] != 150 {
		t.Error("wrong balance")
	}

	// log.Info("satoshi has ", nodestate.cashbalance["satoshi"])
	// log.Info("hal has ", nodestate.cashbalance["hal"])

	// log.Info("satoshi has ", nodestate.cashbalance["satoshi"])
	// log.Info("hal has ", nodestate.cashbalance["hal"])

}
