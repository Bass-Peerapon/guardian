package model

type User struct {
	UserName  string    `json:"username"`
	Roles     []*Role   `json:"roles"`
	CreatedAt Timestamp `json:"created_at"`
	UpdatedAt Timestamp `json:"updated_at"`
}
