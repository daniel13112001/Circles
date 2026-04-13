package groups

import "errors"

var (
	ErrNotFound      = errors.New("group not found")
	ErrForbidden     = errors.New("forbidden")
	ErrNotFriend     = errors.New("user is not your friend")
	ErrCircleCheck   = errors.New("user is not friends with all current members")
	ErrAlreadyMember = errors.New("user is already a member")
)
