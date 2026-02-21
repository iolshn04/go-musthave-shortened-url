package repository

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("url not found")
var ErrDeleted = errors.New("url deleted")

type ErrAlreadyExistsWithID struct {
	ExistingID string
}

func (e ErrAlreadyExistsWithID) Error() string {
	return fmt.Sprintf("already exists with id %s", e.ExistingID)
}
