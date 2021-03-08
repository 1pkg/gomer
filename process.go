package main

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	sortVersion = iota + 1
	sortTimestamp
)

const (
	printVersion = iota + 1
	printTimestamp
)

func process(
	ctx context.Context,
	ichan <-chan modv,
	r *regexp.Regexp,
	sorter int8,
	printer int8,
) error {
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
		switch sorter {
		case sortTimestamp:
			return mods[i].Timestamp.Before(mods[i].Timestamp)
		case sortVersion:
			return strings.Compare(mods[i].Version, mods[j].Version) == 1
		default:
			return false
		}
	})
	for _, mod := range mods {
		fmt.Print(mod.Path)
		switch printer {
		case printVersion:
			fmt.Print(" ", mod.Version)
		case printTimestamp:
			fmt.Print(" ", mod.Timestamp.Format(time.RFC3339))
		case printVersion | printTimestamp:
			fmt.Print(" ", mod.Version, " ", mod.Timestamp.Format(time.RFC3339))
		}
		fmt.Print("\n")
	}
	return nil
}
