package postgres

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound      = errors.New("entity not found")
	ErrNotModified   = errors.New("no rows modified")
	ErrAlreadyExists = errors.New("entity already exists")
)

func formatError(queryName string, err error) error {
	return fmt.Errorf("executing %s: %w", queryName, err)
}
