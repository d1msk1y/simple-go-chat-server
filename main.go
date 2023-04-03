package main

import (
	"encoding/json"
	"fmt"
	"github.com/d1msk1y/simple-go-chat-server/limiter"
	"github.com/d1msk1y/simple-go-chat-server/models"
	"github.com/d1msk1y/simple-go-chat-server/pagination"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

var conn *websocket.Conn

// test slice of message structs
var messages = []models.Message{
	{ID: "0", Username: "d1msk1y 1", Time: "00:00", Message: "Hellow World!"},
	{ID: "1", Username: "d1msk1y 2", Time: "00:01", Message: "Hellow d1msk1y!"},
	{ID: "2", Username: "d1msk1y 1", Time: "00:02", Message: "How ya doin?"},
	{ID: "3", Username: "d1msk1y 2", Time: "00:03", Message: "aight, and u?"},
}
var pageSize int = 10

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

	limiterInstance := limiter.GetLimiter()

	router.Use(func(c *gin.Context) {
		limiterContext, err := limiterInstance.Get(c.Request.Context(), "rate-limit")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error occurred while checking the rate limit",
			})
			c.Abort()
			return
		}
		if limiterContext.Reached {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			c.Abort()
			return
		}

		c.Next()
	})

	router.GET("/messages", getMessagesByPage)
	router.GET("/messages/:id", getMessageByID)
	router.GET("/messages/pages/:page", getMessagesByPage)
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
	var newMessage models.Message

	if err := c.BindJSON(&newMessage); err != nil {
		return
	}
	messages = append(messages, newMessage)
	c.IndentedJSON(http.StatusCreated, newMessage)

	messageJson, _ := json.Marshal(messages[len(messages)-1])
	conn.WriteMessage(websocket.TextMessage, messageJson)
}

func getMessagesByPage(c *gin.Context) {
	pageId := c.Param("page")
	messages := messages

	paginatedMessages := pagination.Paginate(messages, pageSize, pageId, c)

	c.IndentedJSON(http.StatusOK, gin.H{
		"messages": paginatedMessages,
		"pageSize": pageSize,
		"total":    len(messages),
	})
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

func getLastMessagePage(c *gin.Context) {

	lastPageId := int(float64(len(messages)/pageSize) + 0.5)
	paginatedMessages := pagination.Paginate(messages, pageSize, string(lastPageId), c)

	c.IndentedJSON(http.StatusOK, gin.H{
		"messages": paginatedMessages,
		"pageSize": pageSize,
		"total":    len(messages),
	})
}
