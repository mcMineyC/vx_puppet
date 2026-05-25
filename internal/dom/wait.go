package dom

import (
	"context"

	"github.com/chromedp/chromedp"
)

func WaitVisible(
	ctx context.Context,
	selector string,
) error {

	return chromedp.Run(
		ctx,
		chromedp.WaitVisible(
			selector,
			chromedp.ByQuery,
		),
	)
}
