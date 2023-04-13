package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/d1msk1y/simple-go-chat-server/limiter"
	"github.com/d1msk1y/simple-go-chat-server/models"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"strconv"
)

var conn *websocket.Conn
var db *sql.DB

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
	cfg := mysql.Config{
		User:   os.Getenv("DBUSER"),
		Passwd: os.Getenv("DBPASS"),
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "chat",
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")
	//db.Exec("INSERT INTO Messages (username, time, message) VALUES (?, ?, ?)", messages[1].Username, messages[1].Time, messages[1].Message)
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
	router.GET("/messages/all", getAllMessages)
	router.GET("/messages/:id", getMessageByID)
	router.GET("/messages/pages/:page", getMessagesByPage)
	router.GET("/messages/pages/last", getLastMessagePage)
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
		fmt.Println(err)
		return
	}

	result, err := db.Exec("INSERT INTO Messages (username, time, message) VALUES (?, ?, ?)",
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
	conn.WriteMessage(websocket.TextMessage, messageJson)
}

func getMessagesByPage(c *gin.Context) {
	pageId := c.Param("page")

	var messages []models.Message

	parsedId, err := strconv.ParseInt(pageId, 6, 12)
	startOffset := parsedId * 10

	rows, err := db.Query("SELECT * FROM Messages ORDER BY ID DESC LIMIT ? OFFSET ?", pageSize, startOffset)
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

	row := db.QueryRow("SELECT * FROM Messages WHERE id = ?", id)

	var message models.Message
	if err := row.Scan(&message.ID, &message.Username, &message.Time, &message.Message); err != nil {
		if err == sql.ErrNoRows {
			fmt.Errorf("albumsById %d: no such album")
		}
		fmt.Errorf("messagesFromDB %q: %v", err)
	}
	c.IndentedJSON(http.StatusOK, message)
}

// test func
func getAllMessages(c *gin.Context) {
	var messages []models.Message

	rows, err := db.Query("SELECT * FROM Messages")
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

func getLastMessage(c *gin.Context) {
	row := db.QueryRow("SELECT * FROM Messages WHERE id = (SELECT MAX(id) FROM Messages)")

	var message models.Message
	if err := row.Scan(&message.ID, &message.Username, &message.Time, &message.Message); err != nil {
		if err == sql.ErrNoRows {
			fmt.Errorf("lastMessage %d: nothing found")
		}
		fmt.Errorf("messagesFromDB %q: %v", err)
	}
	c.IndentedJSON(http.StatusOK, message)
}

func getLastMessagePage(c *gin.Context) {
	var messages []models.Message

	rows, err := db.Query("SELECT * FROM Messages ORDER BY ID DESC LIMIT ?", pageSize)
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
