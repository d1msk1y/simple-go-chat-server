package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/d1msk1y/simple-go-chat-server/auth"
	"github.com/d1msk1y/simple-go-chat-server/database"
	"github.com/d1msk1y/simple-go-chat-server/limiter"
	"github.com/d1msk1y/simple-go-chat-server/models"
	multi_room "github.com/d1msk1y/simple-go-chat-server/multi-room"
	"github.com/d1msk1y/simple-go-chat-server/net"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
)

var router = gin.Default()

var pageSize = 10

func main() {
	err := database.TryConnectDB()
	if err != nil {
		return
	}
	runServer()
}

func runServer() {
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

	router.GET("/messages/all", getAllMessages)
	router.GET("/messages/:id", getMessageByID)
	router.GET("/messages/pages/:page", getMessagesByPage)

	router.GET("/rooms/new", multi_room.PostRoom)
	router.GET("/rooms/code/:code", multi_room.GetRoomByCode)
	router.GET("/rooms/id/:id", multi_room.GetRoomByID)

	router.POST("/messages", postMessage)
	router.GET("/auth", auth.TryAuthUser)

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Chat server is running!")
	})

	router.GET("/ws", func(c *gin.Context) {
		net.WSHandler(c.Writer, c.Request)
	})

	err := router.Run("localhost:8080")
	if err != nil {
		return
	}
}

func postMessage(c *gin.Context) {
	var newMessage models.Message
	if err := c.BindJSON(&newMessage); err != nil {
		fmt.Println(err)
		return
	}

	// Assign message to a specific room
	result, err := database.DB.Exec("INSERT INTO Messages (username, time, message) VALUES (?, ?, ?)",
		newMessage.Username,
		newMessage.Time,
		newMessage.Message)
	if err != nil {
		fmt.Errorf("addMessage ", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Inserted %d rows into the Messages table\n", rowsAffected)

	c.IndentedJSON(http.StatusCreated, newMessage)

	messageJson, _ := json.Marshal(newMessage)
	net.Conn.WriteMessage(websocket.TextMessage, messageJson)
}

func getMessagesByPage(c *gin.Context) {
	pageId := c.Param("page")
	var messages []models.Message
	roomId := multi_room.GetRoomID(c)

	parsedId, err := strconv.ParseInt(pageId, 6, 12)
	startOffset := parsedId * 10

	rows, err := database.DB.Query("SELECT * FROM Messages Where roomd_id = ? ORDER BY ID DESC LIMIT ? OFFSET ?", roomId, pageSize, startOffset)
	if err != nil {
		fmt.Errorf("messagesFromDB %q: %v", err)
	}

	defer rows.Close()
	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.ID, &message.Username, &message.Time, &message.Message); err != nil {
			fmt.Errorf("messagesFromDB %q: %v", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		fmt.Errorf("messagesFromDB %q: %v", err)
	}
	c.IndentedJSON(http.StatusOK, gin.H{
		"messages": messages,
		"pageSize": pageSize,
		"total":    len(messages),
	})
}

func getMessageByID(c *gin.Context) {
	id := c.Param("id")
	roomId := multi_room.GetRoomID(c)

	row := database.DB.QueryRow("SELECT * FROM Messages WHERE room_id = ? ORDER BY ID desc LIMIT ?, 1;", roomId, id)

	var message models.Message
	if err := row.Scan(&message.ID, &message.Username, &message.Time, &message.Message, &message.RoomId); err != nil {
		if err == sql.ErrNoRows {
			fmt.Errorf("messageById %d: no such message")
		}
		fmt.Errorf("messagesFromDB %q: %v", err)
	}
	c.IndentedJSON(http.StatusOK, message)
}

// test func
func getAllMessages(c *gin.Context) {
	var messages []models.Message

	rows, err := database.DB.Query("SELECT * FROM Messages")
	if err != nil {
		fmt.Errorf("messagesFromDB %q: %v", err)
	}

	defer rows.Close()
	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.ID, &message.Username, &message.Time, &message.Message); err != nil {
			fmt.Errorf("messagesFromDB %q: %v", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		fmt.Errorf("messagesFromDB %q: %v", err)
	}
	c.IndentedJSON(http.StatusOK, messages)
}
