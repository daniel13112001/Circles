package posts

import "errors"

var ErrForbidden = errors.New("not a member of this group")
