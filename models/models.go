package models

import "database/sql"

type Message struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Time      string `json:"time"`
	Message   string `json:"message"`
	RoomToken string `json:"roomId"`
}

type Room struct {
	Token string `json:"token"`
}

type User struct {
	ID        string         `json:"id"`
	Username  string         `json:"username"`
	JWT       string         `json:"jwt"`
	RoomToken sql.NullString `json:"room_id"`
}
