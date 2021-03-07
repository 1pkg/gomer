package main

import (
	"context"
	"log"
	"regexp"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"
)

const psize = 2000

type modv struct {
	Path      string
	Version   string
	Timestamp time.Time
}

func main() {
	g, ctx := errgroup.WithContext(context.Background())
	ch := make(chan modv, psize*4)
	defer close(ch)
	t := time.Now().UTC()
	end := time.Date(2019, 04, 10, 19, 8, 52, 997264, time.UTC)
	for t.After(end) {
		g.Go(func() error {
			return fetch(ctx, t.Add(-time.Hour*24*30), t, ch)
		})
		t = t.Add(-time.Hour * 24 * 30)
	}
	for i := 0; i < runtime.NumCPU(); i++ {
		g.Go(func() error {
			return process(ctx, ch, regexp.MustCompile(`gopium`))
		})
	}
	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
