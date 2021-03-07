package main

import (
	"context"
	"regexp"
	"sort"
)

func process(ctx context.Context, r *regexp.Regexp, ichan <-chan modv, p printer) error {
	var mods []modv
loop:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case mod, ok := <-ichan:
			if !ok {
				break loop
			}
			if r.MatchString(mod.Path) {
				mods = append(mods, mod)
			}
		}
	}
	sort.Slice(mods, func(i int, j int) bool {
		return mods[i].Timestamp.After(mods[i].Timestamp)
	})
	for _, mod := range mods {
		p(mod)
	}
	return nil
}
