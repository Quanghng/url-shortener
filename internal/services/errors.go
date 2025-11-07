package services

import "errors"

var (
	ErrShortCodeRequired = errors.New("short code is required")
	ErrLinkNotFound      = errors.New("short link not found")
)
