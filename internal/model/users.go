package model

type User struct {
	ID       int64  `json:"id"`
	UUID     string `json:"uuid"`  // Task3
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}
