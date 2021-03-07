package main

import (
	"context"
	"time"
)

func main() {
	fmonth(context.TODO(), 1, 2020, nil)
}

type modv struct {
	Path      string
	Version   string
	Timestamp time.Time
}
