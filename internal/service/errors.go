package service

import "errors"

var (
	ErrInvalidURL   = errors.New("invalid URL")
	ErrLinkNotFound = errors.New("link not found")
)
