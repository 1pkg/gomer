package main

import (
	"context"
	"time"
)

func fmonth(ctx context.Context, m, y int, mchan chan<- modv) error {
	from, to := time.Date(y, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(y, time.January, 30, 0, 0, 0, 0, time.UTC)
	for {
		mods, t, err := fpage(ctx, from)
		if err != nil {
			return err
		}
		for _, mod := range mods {
			mchan <- mod
		}
		if t.After(to) {
			return nil
		}
		from = t
	}
}

func fpage(ctx context.Context, t time.Time) ([]modv, time.Time, error) {
	return nil, time.Now(), nil
}
