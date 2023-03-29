package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

type message struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Time     string `json:"time"`
	Message  string `json:"message"`
}

var conn *websocket.Conn

// test slice of message structs
var messages = []message{
	{ID: "0", Username: "d1msk1y 1", Time: "00:00", Message: "Hellow World!"},
	{ID: "1", Username: "d1msk1y 2", Time: "00:01", Message: "Hellow d1msk1y!"},
	{ID: "2", Username: "d1msk1y 1", Time: "00:02", Message: "How ya doin?"},
	{ID: "3", Username: "d1msk1y 2", Time: "00:03", Message: "aight, and u?"},
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	runServer()
}

func wshandler(w http.ResponseWriter, r *http.Request) {
	var err error
	conn, err = wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgradeL %+v", err)
		return
	}
}

func runServer() {
	router := gin.Default()
	router.GET("/messages", getMessages)
	router.GET("/messages/:id", getMessageByID)
	router.GET("/messages/last", getLastMessage)
	router.POST("/messages", postMessage)

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Chat server is running!")
	})

	router.GET("/ws", func(c *gin.Context) {
		wshandler(c.Writer, c.Request)
	})

	err := router.Run("localhost:8080")
	if err != nil {
		return
	}
}

func postMessage(c *gin.Context) {
	var newMessage message

	if err := c.BindJSON(&newMessage); err != nil {
		return
	}

	messages = append(messages, newMessage)
	c.IndentedJSON(http.StatusCreated, newMessage)

	conn.WriteMessage(websocket.TextMessage, []byte(messages[len(messages)-1].Message))
}

func getMessages(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, messages)
}

func getMessageByID(c *gin.Context) {
	id := c.Param("id")

	for _, m := range messages {
		if m.ID == id {
			c.IndentedJSON(http.StatusOK, m)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "message not found!"})
}

func getLastMessage(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, messages[len(messages)-1])

}
