package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/d1msk1y/simple-go-chat-server/limiter"
	"github.com/d1msk1y/simple-go-chat-server/models"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var conn *websocket.Conn
var db *sql.DB
var router = gin.Default()

var secretKey = []byte(os.Getenv("CHATSECRET"))

var pageSize = 10

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	err := tryConnectDB()
	if err != nil {
		return
	}
	runServer()
}

func generateJWT(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(10 * time.Minute)
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
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				c.Writer.WriteHeader(http.StatusUnauthorized)
				_, err := c.Writer.Write([]byte("You're Unauthorized!"))
				if err != nil {
					return nil, err
				}
			}
			return "", nil
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

func tryConnectDB() error {
	cfg := mysql.Config{
		User:                 os.Getenv("DBUSER"),
		Passwd:               os.Getenv("DBPASS"),
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "chat",
		AllowNativePasswords: true,
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
		return err
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
		return err
	}
	fmt.Println("Connected!")
	return nil
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
	router.GET("/messages/all", verifyJWT(getAllMessages))
	router.GET("/messages/:id", getMessageByID)
	router.GET("/messages/pages/:page", getMessagesByPage)
	router.GET("/messages/pages/last", getLastMessagePage)
	router.GET("/messages/last", getLastMessage)

	router.POST("/messages", postMessage)
	router.GET("/auth", tryAuthUser)

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

func addNewUser(username string) (string, error) {
	token, err := generateJWT(username)
	if err != nil {
		return "", fmt.Errorf("Error occurred: ", err)
	}

	result, err := db.Exec("INSERT INTO Users (Username, JWT) VALUES (?, ?)",
		username,
		token)
	if err != nil {
		return "", fmt.Errorf("addMessage ", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Inserted %d rows into the Messages table\n", rowsAffected)

	return token, nil
}

func tryAuthUser(c *gin.Context) {
	username := c.GetHeader("Username")

	row := db.QueryRow("SELECT * FROM Users WHERE username = ?;", username)

	var user models.User
	err := row.Scan(&user.Username, &user.JWT)
	if err != nil {
		fmt.Println("userFromDB %q: %v", err)
	}

	if err == sql.ErrNoRows {
		fmt.Println("userById %d: no such user, authorizing the new one...")
		token, err := addNewUser(username)
		if err != nil {
			fmt.Println("couldn't authorize the new user")
		}
		c.IndentedJSON(http.StatusOK, token)
	} else {
		fmt.Println("Authorizing...")
		token := c.GetHeader("Token")
		fmt.Println("User token: ", token)
		c.IndentedJSON(http.StatusOK, "User authorized!")
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
			fmt.Errorf("messageById %d: no such message")
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
