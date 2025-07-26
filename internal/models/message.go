package models

import "time"

type Message struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"userID" db:"user_id"`
	ModelID   int64     `json:"modelID" db:"model_id"`
	VersionID int64     `json:"versionID" db:"version_id"`
	Input1    []byte    `json:"input1" db:"input1"`
	Input2    []byte    `json:"input2" db:"input2"`
	Results   []string  `json:"results" db:"results"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type SendMessageRequest struct {
	Input1 []int32 `json:"input1" example:"0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15"`
	Input2 []int32 `json:"input2" example:"1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16"`
}

type SendMessageResponse struct {
	Results []string `json:"results" example:"1 + 1 = 2,1 - 1 = 0"`
}

type GetMessagesResponse struct {
	Messages []Message `json:"messages"`
}
