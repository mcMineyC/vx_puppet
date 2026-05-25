package vx_puppet

import "time"

type Options struct {
	WebSocketURL string
	SchoolID     string
	Timeout      time.Duration
}
