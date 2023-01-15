package main

import (
	"fmt"
	"time"

	"github.com/horizonledger/protocol"
	log "github.com/sirupsen/logrus"
)

func pubVertexs(pubsub *Pubsub) {
	quit := make(chan struct{})
	//could make this a macro somehow?
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			log.Info("#vertexs ", len(nodestate.vertexs))
			s := fmt.Sprintf("%d", len(nodestate.vertexs))
			msg := protocol.Msg{Type: "vertex", Value: s}
			pubsub.Publish("vertex", protocol.MsgToGen(msg))
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func pubStatus(pubsub *Pubsub) {
	quit := make(chan struct{})
	ticker := time.NewTicker(20 * time.Second)
	for {
		//log.Println("statusLoop")
		select {
		case <-ticker.C:
			log.Debug("status last update: ", nodestate.msgstate.LastUpdate.String())
			//log.Debug(len(vertexs))
			xmsg := protocol.Msg{Type: "STATUS", Value: nodestate.msgstate.LastUpdate.String()}
			pubsub.Publish("status", protocol.MsgToGen(xmsg))
			// for _, v := range vertexs {
			// 	//v.out_write <- protocol.MsgToGen(xmsg)
			// }

		case <-quit:
			ticker.Stop()
			return
		}
	}
}
