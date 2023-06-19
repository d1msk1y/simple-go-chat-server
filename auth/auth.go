package auth

import (
	"database/sql"
	"fmt"
	"github.com/d1msk1y/simple-go-chat-server/database"
	"github.com/d1msk1y/simple-go-chat-server/jwt"
	"github.com/d1msk1y/simple-go-chat-server/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func addNewUser(username string) (string, error) {
	token, err := jwt.GenerateJWT(username)
	if err != nil {
		return "", fmt.Errorf("Error occurred: ", err)
	}

	result, err := database.DB.Exec("INSERT INTO Users (Username, jwt) VALUES (?, ?)",
		username,
		token)
	if err != nil {
		return "", fmt.Errorf("addMessage ", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Inserted %d rows into the Messages table\n", rowsAffected)

	return token, nil
}

func TryAuthUser(c *gin.Context) {
	username := c.GetHeader("Username")

	row := database.DB.QueryRow("SELECT * FROM Users WHERE username = ?;", username)

	var user models.User

	err := row.Scan(&user.Username, &user.JWT, &user.RoomToken)
	if err != nil {
		fmt.Println("userFromDB %q: %v", err)
	}

	if err == sql.ErrNoRows {
		fmt.Println("userById %d: no such user, authorizing the new one...")
		token, err := addNewUser(username)
		if err != nil {
			fmt.Println("couldn't authorize the new user")
		}
		newUser := models.User{
			Username: username,
			JWT:      token,
		}
		c.IndentedJSON(http.StatusOK, newUser)
	} else {
		//token = user.jwt
		c.IndentedJSON(http.StatusOK, user)
	}
}
