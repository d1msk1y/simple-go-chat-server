package multi_room

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/d1msk1y/simple-go-chat-server/database"
	"github.com/d1msk1y/simple-go-chat-server/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

var roomTokenLength int = 10

func GenerateRoomToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func GetRoomID(c *gin.Context) string {
	return c.GetHeader("RoomToken")
}

func GetRoomFromDB(c *gin.Context, query string, args interface{}) models.Room {
	row := database.DB.QueryRow(query, args)
	var room models.Room
	if err := row.Scan(&room.ID, &room.Token); err != nil {
		if err == sql.ErrNoRows {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "No such room"})
			return models.Room{}
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Couldn't get room from DB"})
		return models.Room{}
	}
	return room
}

func PostRoom(c *gin.Context) {
	var roomCode = models.Room{
		Token: GenerateRoomToken(roomTokenLength),
	}

	result, err := database.DB.Exec("INSERT INTO Rooms (code) VALUES (?);",
		roomCode.Token)
	if err != nil {
		fmt.Errorf("addItem ", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Inserted %d rows into the table\n", rowsAffected)

	query := "SELECT * FROM Rooms ORDER BY ID desc LIMIT ?, 1;"
	newRoom := GetRoomFromDB(c, query, 0)
	c.IndentedJSON(http.StatusCreated, newRoom)
}

func GetRoomByCode(c *gin.Context) {
	roomCode := c.GetHeader("RoomCode")
	room := GetRoomFromDB(c, "SELECT * FROM Rooms WHERE code = ?", roomCode)
	c.IndentedJSON(http.StatusCreated, room)
}

func GetRoomByID(c *gin.Context) {
	roomID := c.GetHeader("RoomToken")
	room := GetRoomFromDB(c, "SELECT * FROM Rooms WHERE code = ?", roomID)
	c.IndentedJSON(http.StatusCreated, room)
}

func AssignUserToRoom(c *gin.Context) {
	roomID := c.GetHeader("RoomToken")
	username := c.GetHeader("Username")
	result, err := database.DB.Exec("UPDATE Users SET room_id=? WHERE username=?", roomID, username)
	if err != nil {
		fmt.Println("assign user: ", err)
	}
	fmt.Println(result)
}
