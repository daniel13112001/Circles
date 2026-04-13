package contacts

import "errors"

var (
	ErrNotFound  = errors.New("contact not found")
	ErrDuplicate = errors.New("contact already exists")
	ErrSelfAdd   = errors.New("cannot add your own phone hash")
)
