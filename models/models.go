package models

type Message struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Time     string `json:"time"`
	Message  string `json:"message"`
	RoomId   string `json:"roomId"`
}

type Room struct {
	ID   string `json:"id"`
	Code string `json:"code"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	JWT      string `json:"jwt"`
}
