package posts

import "time"

type Post struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"group_id"`
	Author    Author    `json:"author"`
	ImageURL  string    `json:"image_url"`
	Caption   *string   `json:"caption"`
	CreatedAt time.Time `json:"created_at"`
}

type Author struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}
