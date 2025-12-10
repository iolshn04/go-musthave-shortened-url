package repository

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("url not found")

type ErrAlreadyExistsWithID struct {
	ExistingID string
}

func (e ErrAlreadyExistsWithID) Error() string {
	return fmt.Sprintf("already exists with id %s", e.ExistingID)
}
