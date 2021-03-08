package main

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

func process(ctx context.Context, ichan <-chan modv, r *regexp.Regexp, format string) error {
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
		if cmp := strings.Compare(mods[i].Path, mods[j].Path); cmp != 0 {
			return cmp == -1
		}
		if cmp := strings.Compare(mods[i].Version, mods[j].Version); cmp != 0 {
			return cmp == 1
		}
		return mods[i].Timestamp.Before(mods[i].Timestamp)
	})
	for _, mod := range mods {
		fmt.Printf(format+"\n", mod.Path, mod.Version, mod.Timestamp.Format(time.RFC3339Nano))
	}
	return nil
}
