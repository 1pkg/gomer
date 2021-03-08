package main

import (
	"context"
	"flag"
	"log"
	"regexp"
	"time"
)

type modv struct {
	Path      string
	Version   string
	Timestamp time.Time
}

func main() {
	cached := flag.Bool("-cached", false, "")
	format := flag.String("-format", "%s %s %s\n", "")
	timeout := flag.Int64("-timeout", 0, "")
	flag.Parse()
	name := flag.Arg(0)
	r, err := regexp.Compile(name)
	if err != nil {
		log.Fatal(r)
	}
	ctx := context.Background()
	if t := *timeout; t > 0 {
		tctx, cancel := context.WithTimeout(ctx, time.Duration(t))
		defer cancel()
		ctx = tctx
	}
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()
	ch := fetch(ctx, *cached)
	if err := process(ctx, ch, r, *format); err != nil {
		log.Fatal(err)
	}
}
