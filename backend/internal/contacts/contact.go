package contacts

import "time"

type Contact struct {
	ID          string    `json:"id"`
	PhoneHash   string    `json:"phone_hash"`
	DisplayName *string   `json:"display_name"` // nil if not yet matched
	Matched     bool      `json:"matched"`
	CreatedAt   time.Time `json:"created_at"`
}
