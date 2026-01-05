package repository

import "errors"

// ErrNotFound is returned when the requested entity does not exist.
var ErrNotFound = errors.New("not found")
