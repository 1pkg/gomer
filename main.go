package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/Masterminds/semver/v3"
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
			"Usage of gomer: [-f val] [-f val] <module_path_regexp> \n",
		)
		flag.PrintDefaults()
	}
	constraint := flag.String("constraint", "", "cli semver constraint pattern; used only if non empty valid constraint specified")
	format := flag.String("format", "%s %s %s", "cli printf format for printing a module entry; \\n is auto appended")
	index := flag.String("index", "https://index.golang.org/index", "cli go module index database url; golang.org index is used by default")
	nocache := flag.Bool("nocache", false, "cli modules no caching flag; caching is enabled by default (default false)")
	timeout := flag.Int64("timeout", 0, "cli timeout in seconds; only considered when value bigger than 0 (default 0)")
	verbose := flag.Bool("verbose", false, "cli verbosity logging flag; verbosity is disabled by default (default false)")
	flag.Parse()
	name := flag.Arg(0)
	r, err := regexp.Compile(name)
	if err != nil {
		log.Fatal(r)
	}
	var c *semver.Constraints
	if *constraint != "" {
		c, err = semver.NewConstraint(*constraint)
		if err != nil {
			log.Fatal(r)
		}
	}
	ctx := context.Background()
	if t := *timeout; t > 0 {
		tctx, cancel := context.WithTimeout(ctx, time.Duration(t)*time.Second)
		defer cancel()
		ctx = tctx
	}
	ch := fetch(ctx, *index, !*nocache, *verbose)
	if err := process(ctx, ch, r, c, *format); err != nil {
		log.Fatal(err)
	}
}
