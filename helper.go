package vx_puppet

import (
	"context"
	"log"
	"time"
)

func WithDebugTimeout(
	parent context.Context,
	name string,
	timeout time.Duration,
) (context.Context, context.CancelFunc) {

	ctx, cancel := context.WithTimeout(
		parent,
		timeout,
	)

	go func() {
		<-ctx.Done()

		log.Printf(
			"[CTX DONE] %s: %v\n",
			name,
			ctx.Err(),
		)
	}()

	return ctx, func() {

		log.Printf(
			"[CTX CANCEL] %s\n",
			name,
		)

		cancel()
	}
}
