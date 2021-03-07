package main

import (
	"context"
	"log"
	"regexp"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	psize    = 2000
	fwindow  = time.Hour * 24 * 30
	mretries = 4
)

var origin = time.Date(2019, 04, 10, 19, 8, 52, 997264, time.UTC)

type modv struct {
	Path      string
	Version   string
	Timestamp time.Time
}

func main() {
	gf, ctx := errgroup.WithContext(context.Background())
	gp, ctx := errgroup.WithContext(context.Background())
	ch := make(chan modv, psize*4)
	t := time.Now().UTC()
	for i := 0; i < runtime.NumCPU(); i++ {
		gp.Go(func() error {
			return process(ctx, ch, regexp.MustCompile(`gopium`))
		})
	}
	for t.After(origin) {
		from, to := t.Add(-fwindow), t
		if from.Before(origin) {
			from = origin
		}
		gf.Go(func() error {
			return fetch(ctx, from, to, ch)
		})
		t = from
	}
	if err := gf.Wait(); err != nil {
		log.Fatal(err)
	}
	close(ch)
	if err := gp.Wait(); err != nil {
		log.Fatal(err)
	}
}
