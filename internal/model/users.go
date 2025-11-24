package model

type User struct {
	ID       int64  `json:"id"`
	UUID     string `json:"uuid"`                           // Task3
	Username string `json:"username" binding:"required"`    // Task4: validation added
	Email    string `json:"email" binding:"required,email"` // Task4: validation added
	FullName string `json:"full_name"`
}
