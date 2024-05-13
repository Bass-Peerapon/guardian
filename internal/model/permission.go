package model

type Application struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Permission struct {
	ID          string    `json:"id"`
	AppID       string    `json:"app_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   Timestamp `json:"created_at"`
}

type Role struct {
	ID          string        `json:"id"`
	AppID       string        `json:"app_id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Permissions []*Permission `json:"permissions"`
	CreatedAt   Timestamp     `json:"created_at"`
}
