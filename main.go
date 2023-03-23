package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type message struct {
	Username string `json:"username"`
	Time     string `json:"time"`
	Message  string `json:"message"`
}

// test slice of message structs
var messages = []message{
	{Username: "d1msk1y 1", Time: "00:00", Message: "Hellow World!"},
	{Username: "d1msk1y 2", Time: "00:01", Message: "Hellow d1msk1y!"},
	{Username: "d1msk1y 1", Time: "00:02", Message: "How ya doin?"},
	{Username: "d1msk1y 2", Time: "00:03", Message: "aight, and u?"},
}

func main() {
	router := gin.Default()
	router.GET("/messages", getMessages)
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Chat server is running!")
	})

	err := router.Run("localhost:8080")
	if err != nil {
		return
	}
}

func getMessages(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, messages)
}
