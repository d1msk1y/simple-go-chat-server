package net

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

var Conn *websocket.Conn

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	Conn, err = wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgradeL %+v", err)
		return
	}
}
