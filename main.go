package main

import (
	"context"
	"log"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	g, ctx := errgroup.WithContext(context.Background())
	ch := make(chan modv, 8000)
	cy := time.Now().Year()
	for y := 2019; y <= cy; y++ {
		for m := 1; m <= 12; m++ {
			g.Go(func() error {
				return fmonth(ctx, y, m, ch)
			})
		}
	}
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}

type modv struct {
	Path      string
	Version   string
	Timestamp time.Time
}
