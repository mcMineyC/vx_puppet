package vx_puppet

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type Client struct {
	allocatorCtx context.Context
	tabCtx       context.Context
	cancel       context.CancelFunc

	baseURL string
	timeout time.Duration
}

func New(opts Options) (*Client, error) {

	if opts.WebSocketURL == "" {
		return nil, ErrMissingWebsocketURL
	}

	baseURL := "https://accounts.veracross.com/" + opts.SchoolID

	timeout := 90 * time.Second

	if opts.Timeout > 0 {
		timeout = opts.Timeout
	}

	allocatorCtx, _ := chromedp.NewRemoteAllocator(
		context.Background(),
		opts.WebSocketURL,
	)

	tabCtx, cancel := chromedp.NewContext(
		allocatorCtx,
	)

	return &Client{
		allocatorCtx: allocatorCtx,
		tabCtx:       tabCtx,
		cancel:       cancel,
		baseURL:      baseURL,
		timeout:      timeout,
	}, nil
}

func (c *Client) Close() {
	c.cancel()
}

func (c *Client) Login(username, password string) error {

	loginURL := c.baseURL + "/portals/login"

	err := chromedp.Run(c.tabCtx,
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(`form#login-form`),
		chromedp.WaitVisible(`#username`),
		chromedp.WaitVisible(`#password`),
	)
	if err != nil {
		return err
	}

	err = chromedp.Run(c.tabCtx,
		chromedp.SetValue(`#username`, username),
		chromedp.SetValue(`#password`, password),
	)
	if err != nil {
		return err
	}

	err = chromedp.Run(c.tabCtx,
		chromedp.Evaluate(`document.querySelector('form').submit()`, nil),
	)

	// IMPORTANT: ignore navigation interruption
	if err != nil &&
		!strings.Contains(err.Error(), "context canceled") {
		return err
	}

	successCh := make(chan error, 1)
	warningCh := make(chan error, 1)

	go func() {
		waitCtx, cancel := context.WithTimeout(c.tabCtx, 30*time.Second)
		defer cancel()

		successCh <- chromedp.Run(waitCtx,
			chromedp.WaitVisible(`.component-class-list-student`),
		)
	}()

	go func() {
		waitCtx, cancel := context.WithTimeout(c.tabCtx, 30*time.Second)
		defer cancel()
		var warning string

		chromedp.Run(waitCtx,
			chromedp.WaitVisible(`.vx-MessageBanner.warning`),
			chromedp.InnerHTML(".vx-MessageBanner.warning", &warning, chromedp.ByQuery),
		)
		warningCh <- fmt.Errorf("login warning: %s", warning)
	}()

	select {
	case err := <-successCh:
		return err
	case err := <-warningCh:
		return err
	}
}

func (c *Client) GetGrades() ([]GradeInfo, error) {

	var ctx = c.tabCtx
	var result []GradeInfo

	err := chromedp.Run(ctx,
		// IMPORTANT: stabilize SPA first
		chromedp.WaitVisible(`body`, chromedp.ByQuery),

		chromedp.WaitReady(".course", chromedp.ByQuery),
		// chromedp.Sleep(2*time.Second),

		// now wait for actual app component
		chromedp.WaitVisible(`.component-class-list-student`, chromedp.ByQuery),
	)
	if err != nil {
		return nil, err
	}

	err = chromedp.Run(ctx,

		chromedp.Evaluate(`
(() => {

	const grades = [];

	const rows = document.querySelectorAll('.course');

	for (const row of rows) {

		const letterBox = row.querySelector('.letter-grade');
		if (!letterBox) continue;

		const className = row.querySelector('a.course-name')?.innerText || '';
		const gradeLetter = letterBox.textContent.trim();
		const grade = row.querySelector('.numeric-grade')?.textContent?.trim() || '';

		let updates = 0;

		const badge = row.querySelector('.notifications-link > .notification-badge');
		if (badge) {
			const num = parseInt(badge.textContent.trim(), 10);
			if (!isNaN(num)) updates = num;
		}

		grades.push({
			class: className,
			gradeLetter,
			grade,
			new_updates: updates
		});
	}

	return grades;

})()
		`, &result),
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}
func (c *Client) attach() error {
	newCtx, cancel := chromedp.NewContext(c.allocatorCtx)

	// verify it's alive
	err := chromedp.Run(newCtx,
		chromedp.Evaluate(`1+1`, nil),
	)
	if err != nil {
		cancel()
		return err
	}

	c.cancel = cancel
	c.tabCtx = newCtx

	return nil
}
