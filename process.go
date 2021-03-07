package main

import (
	"context"
	"fmt"
	"regexp"
)

func process(ctx context.Context, mchan <-chan modv, reg *regexp.Regexp) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case mod, ok := <-mchan:
			if !ok {
				return nil
			}
			if reg.MatchString(mod.Path) {
				fmt.Printf("%v", mod)
			}
		}
	}
}
