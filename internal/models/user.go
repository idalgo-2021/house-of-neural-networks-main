package models

type User struct {
	ID       int64  `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
	Email    string `json:"email" db:"email"`
}

type SignUpRequest struct {
	Username string `json:"username" example:"john_doe"`
	Password string `json:"password" example:"securepassword123"`
	Email    string `json:"email" example:"john@example.com"`
}

type SignUpResponse struct {
	Success bool `json:"success"`
}

type LogInRequest struct {
	Username string `json:"username" example:"john_doe"`
	Password string `json:"password" example:"securepassword123"`
}

type LogInResponse struct {
	Jwt    string `json:"jwt"`
	UserId int64  `json:"userId"`
}
