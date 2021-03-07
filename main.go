package main

import (
	"context"
	"log"
	"regexp"
	"time"

	"golang.org/x/sync/errgroup"
)

type modv struct {
	Path      string
	Version   string
	Timestamp time.Time
}

func main() {
	gp, ctx := errgroup.WithContext(context.Background())
	ch := make(chan modv, bigPageSize)
	f := fetcherParallel{cache: true}
	gp.Go(func() error {
		return process(ctx, regexp.MustCompile(`cobra`), ch, printf(true, false))
	})
	if err := f.fetch(ctx, ch); err != nil {
		log.Fatal(err)
	}
	if err := gp.Wait(); err != nil {
		log.Fatal(err)
	}
}
