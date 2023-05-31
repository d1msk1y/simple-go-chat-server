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

	row := database.DB.QueryRow("SELECT * FROM Rooms ORDER BY ID desc LIMIT ?, 1;", 0)
	var newRoom models.Room
	if err := row.Scan(&newRoom.ID, &newRoom.Code); err != nil {
		if err == sql.ErrNoRows {
			fmt.Errorf("no such room")
		}
		fmt.Errorf("roomFromDB %q: %v", err)
	}
	c.IndentedJSON(http.StatusCreated, newRoom)
}
