package vx_puppet

import "errors"

var (
	ErrMissingWebsocketURL = errors.New("missing websocket url")
	ErrLoginFailed         = errors.New("login failed")
	ErrLoginTimeout        = errors.New("operation timeout")
)
