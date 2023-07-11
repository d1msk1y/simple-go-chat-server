package multi_room

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/d1msk1y/simple-go-chat-server/database"
	"github.com/d1msk1y/simple-go-chat-server/models"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"time"
)

var roomTokenLength int = 10

func GenerateRoomToken(length int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func GetRoomToken(c *gin.Context) string {
	return c.GetHeader("RoomToken")
}

func GetRoomFromDB(c *gin.Context, query string, args interface{}) (models.Room, int) {
	row := database.DB.QueryRow(query, args)
	var room models.Room
	if err := row.Scan(&room.Token); err != nil {
		if err == sql.ErrNoRows {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "No such room"})
			return models.Room{}, http.StatusNotFound
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Couldn't get room from DB"})
		return models.Room{}, http.StatusInternalServerError
	}
	return room, http.StatusOK
}

func PostRoom(c *gin.Context) {
	fmt.Println(GenerateRoomToken(10))
	var newRoom = models.Room{
		Token: GenerateRoomToken(roomTokenLength),
	}

	result, err := database.DB.Exec("INSERT INTO Rooms (token) VALUES (?);",
		newRoom.Token)
	if err != nil {
		fmt.Errorf("addItem ", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Inserted %d rows into the table\n", rowsAffected)

	c.IndentedJSON(http.StatusOK, newRoom)
}

func GetRoomByToken(c *gin.Context) {
	roomToken := c.Param("token")
	room, code := GetRoomFromDB(c, "SELECT * FROM Rooms WHERE token = ?", roomToken)
	c.IndentedJSON(code, room)
}

func GetRoomUsers(c *gin.Context) {
	token := c.Param("token")
	var users []models.User
	row, err := database.DB.Query("SELECT * FROM Users WHERE room_token = ?;", token)
	if err != nil {
		fmt.Println("User sql query", err)
		c.IndentedJSON(http.StatusInternalServerError, err)
		return
	}
	for row.Next() {
		var user models.User
		if err := row.Scan(&user.Username, &user.JWT, &user.RoomToken); err != nil {
			fmt.Println("User sql scan: ", err)
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		users = append(users, user)
	}

	if err := row.Err(); err != nil {
		fmt.Println("User sql row: ", err)
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"users": users,
		"size":  len(users),
	})
}

func AssignUserToRoom(c *gin.Context) {
	roomToken := c.GetHeader("RoomToken")
	username := c.GetHeader("Username")
	result, err := database.DB.Exec("UPDATE Users SET room_token=? WHERE username=?", roomToken, username)
	if err != nil {
		fmt.Println("assign user: ", err)
	}
	fmt.Println(result)
}
