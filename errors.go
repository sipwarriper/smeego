package smeego

import "errors"

var (
	ErrFailedToCreateChannel = errors.New("failed to create channel")
	ErrFailedToSubscribe     = errors.New("failed to subscribe")
)
