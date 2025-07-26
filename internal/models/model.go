package models

type Model struct {
	ID       int64      `json:"id" db:"id"`
	Name     string     `json:"name" db:"name"`
	UserID   int64      `json:"user_id" db:"user_id"`
	Versions []*Version `json:"versions"`
}

type GetModelResponse struct {
	Model Model `json:"model"`
}

type ListModelsResponse struct {
	Models []Model `json:"models"`
}

type UploadModelResponse struct {
	Id int64 `json:"id"`
}

type UploadVersionResponse struct {
	Id int64 `json:"id"`
}

type UnloadModelRequest struct {
	Id int64 `json:"id"`
}

type UnloadModelResponse struct {
	Success bool `json:"success"`
}
