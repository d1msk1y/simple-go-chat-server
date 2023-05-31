package multi_room

import (
	"crypto/rand"
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

func PostRoom(c *gin.Context) {
	var newRoom models.Room = models.Room{
		Code: GenerateRoomToken(roomTokenLength),
	}

	result, err := database.DB.Exec("INSERT INTO Rooms (code) VALUES (?);",
		newRoom.Code)
	if err != nil {
		fmt.Errorf("addMessage ", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Inserted %d rows into the Messages table\n", rowsAffected)

	c.IndentedJSON(http.StatusCreated, newRoom)
}
