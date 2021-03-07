package main

import (
	"fmt"
	"time"
)

type printer func(modv)

func printf(v, t bool) func(modv) {
	return func(m modv) {
		fmt.Print(m.Path)
		if v {
			fmt.Print(" ", m.Version)
		}
		if t {
			fmt.Print(" ", m.Timestamp.Format(time.RFC3339))
		}
		fmt.Print("\n")
	}
}
