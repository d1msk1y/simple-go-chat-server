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
	return c.GetHeader("RoomID")
}

func GetRoomFromDB(query string, args interface{}) (models.Room, error) {
	row := database.DB.QueryRow(query, args)
	var room models.Room
	if err := row.Scan(&room.ID, &room.Code); err != nil {
		if err == sql.ErrNoRows {
			fmt.Errorf("no such room")
		}
		fmt.Errorf("roomFromDB %q: %v", err)
	}
	return room, nil
}

func PostRoom(c *gin.Context) {
	var roomCode = models.Room{
		Code: GenerateRoomToken(roomTokenLength),
	}

	result, err := database.DB.Exec("INSERT INTO Rooms (code) VALUES (?);",
		roomCode.Code)
	if err != nil {
		fmt.Errorf("addItem ", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Inserted %d rows into the table\n", rowsAffected)

	query := "SELECT * FROM Rooms ORDER BY ID desc LIMIT ?, 1;"
	newRoom, err := GetRoomFromDB(query, 0)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error occurred"})
		return
	}

	c.IndentedJSON(http.StatusCreated, newRoom)
}

func GetRoomByCode(c *gin.Context) {
	roomCode := c.GetHeader("RoomCode")

	room, err := GetRoomFromDB("SELECT * FROM Rooms WHERE code = ?", roomCode)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error occurred"})
		return
	}
	c.IndentedJSON(http.StatusCreated, room)
}
