package users

import "time"

type User struct {
	ID          string    `json:"id"`
	FirebaseUID string    `json:"-"`
	DisplayName string    `json:"display_name"`
	PhoneHash   string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}
