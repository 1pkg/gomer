package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"
)

type modv struct {
	Path      string
	Version   string
	Timestamp time.Time
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"Usage of %s: <module_path_regexp> \n",
			os.Args[0],
		)
		flag.PrintDefaults()
	}
	nocache := flag.Bool("nocache", false, "cli modules caching flag; caching is enabled by default (default false)")
	format := flag.String("format", "%s %s %s", "cli printf format for printing a module entry; \\n is auto appended")
	timeout := flag.Int64("timeout", 0, "cli timeout in seconds; only considered when value bigger than 0 (default 0)")
	flag.Parse()
	name := flag.Arg(0)
	r, err := regexp.Compile(name)
	if err != nil {
		log.Fatal(r)
	}
	ctx := context.Background()
	if t := *timeout; t > 0 {
		tctx, cancel := context.WithTimeout(ctx, time.Duration(t)*time.Second)
		defer cancel()
		ctx = tctx
	}
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
		}
	}()
	ch := fetch(ctx, !*nocache)
	if err := process(ctx, ch, r, *format); err != nil {
		log.Fatal(err)
	}
}
