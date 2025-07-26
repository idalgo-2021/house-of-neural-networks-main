package models

type Version struct {
	ID      int64 `json:"id" db:"id"`
	Number  int32 `json:"number" db:"number"`
	ModelID int64 `json:"model_id" db:"model_id"`
}
