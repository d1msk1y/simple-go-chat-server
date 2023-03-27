package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type message struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Time     string `json:"time"`
	Message  string `json:"message"`
}

// test slice of message structs
var messages = []message{
	{ID: "0", Username: "d1msk1y 1", Time: "00:00", Message: "Hellow World!"},
	{ID: "1", Username: "d1msk1y 2", Time: "00:01", Message: "Hellow d1msk1y!"},
	{ID: "2", Username: "d1msk1y 1", Time: "00:02", Message: "How ya doin?"},
	{ID: "3", Username: "d1msk1y 2", Time: "00:03", Message: "aight, and u?"},
}

func main() {
	router := gin.Default()
	router.GET("/messages", getMessages)
	router.GET("/messages/:id", getAlbumByID)
	router.POST("/messages", postMessage)

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Chat server is running!")
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
}

func getMessages(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, messages)
}

func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	for _, a := range messages {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "message not found!"})
}
