package main

import (
	"encoding/json"
	"fmt"
	"github.com/d1msk1y/simple-go-chat-server/auth"
	"github.com/d1msk1y/simple-go-chat-server/database"
	"github.com/d1msk1y/simple-go-chat-server/limiter"
	"github.com/d1msk1y/simple-go-chat-server/models"
	multi_room "github.com/d1msk1y/simple-go-chat-server/multi-room"
	"github.com/d1msk1y/simple-go-chat-server/net"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"strconv"
)

var router = gin.Default()

var secretKey = []byte(os.Getenv("CHATSECRET"))

var pageSize = 10

func main() {
	err := database.TryConnectDB()
	if err != nil {
		return
	}
	runServer()
}

func generateJWT(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["user"] = username

	fmt.Println("TOKEN:", token)

	tokenString, err := token.SignedString(secretKey)
	fmt.Println("TOKENSTRING: ", tokenString)
	if err != nil {
		return "Error occurred while signing JWT", err
	}

	return tokenString, nil
}

func verifyJWT(endpointHandler func(c *gin.Context)) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		tokenString := c.GetHeader("Token")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}
		fmt.Println("Token: ", tokenString)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		})

		if err != nil {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			_, err := c.Writer.Write([]byte("You're Unauthorized!"))
			if err != nil {
				return
			}
		}

		if token.Valid {
			endpointHandler(c)
		} else {
			c.Writer.WriteHeader(http.StatusUnauthorized)
			_, err := c.Writer.Write([]byte("You're Unauthorized due to invalid token"))
			if err != nil {
				return
			}
		}
	})
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

	router.GET("/messages", getMessagesByPage)
	router.GET("/messages/all", getAllMessages)
	router.GET("/messages/:id", getMessageByID)
	router.GET("/messages/pages/:page", getMessagesByPage)

	router.GET("/rooms/new", multi_room.PostRoom)
	router.GET("/rooms/token/:token", multi_room.GetRoomByToken)
	router.GET("/rooms/users/:token", multi_room.GetRoomUsers)
	router.POST("/rooms/join", multi_room.AssignUserToRoom)

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
	result, err := database.DB.Exec("INSERT INTO Messages (username, time, message, room_token) VALUES (?, ?, ?, ?)",
		newMessage.Username,
		newMessage.Time,
		newMessage.Message,
		newMessage.RoomToken)
	if err != nil {
		fmt.Errorf("addMessage ", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Inserted %d rows into the Messages table\n", rowsAffected)

	c.IndentedJSON(http.StatusCreated, newMessage)

	messageJson, _ := json.Marshal(newMessage)
	net.Conn.WriteMessage(websocket.TextMessage, messageJson)
}

func getMessagesFromDB(query string, args ...interface{}) ([]models.Message, error) {
	var messages []models.Message

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		fmt.Println("sql query", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.ID, &message.Username, &message.Time, &message.Message, &message.RoomToken); err != nil {
			fmt.Println("sql scan: ", err)
			return nil, err
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		fmt.Println("sql row: ", err)
		return nil, err
	}

	return messages, nil
}

func getMessagesByPage(c *gin.Context) {
	pageId := c.Param("page")
	roomToken := multi_room.GetRoomToken(c)

	parsedId, _ := strconv.ParseInt(pageId, 6, 12)
	startOffset := parsedId * 10

	messages, _ := getMessagesFromDB("SELECT * FROM Messages Where room_token = ? ORDER BY ID ASC LIMIT ? OFFSET ?", roomToken, pageSize, startOffset)

	c.IndentedJSON(http.StatusOK, gin.H{
		"messages": messages,
		"pageSize": pageSize,
		"total":    len(messages),
	})
}

func getMessageByID(c *gin.Context) {
	id := c.Param("id")
	roomToken := multi_room.GetRoomToken(c)
	message, _ := getMessagesFromDB("SELECT * FROM Messages WHERE room_token = ? ORDER BY ID desc LIMIT ?, 1;", roomToken, id)
	c.IndentedJSON(http.StatusOK, message[0])
}

func getAllMessages(c *gin.Context) {
	messages, _ := getMessagesFromDB("SELECT * FROM Messages ORDER BY ID desc")
	c.IndentedJSON(http.StatusOK, messages)
}
