package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func fmonth(ctx context.Context, y, m int, mchan chan<- modv) error {
	from := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(y, time.Month(m), ndays(y, m), 0, 0, 0, 0, time.UTC)
	for {
		mods, err := fpage(ctx, from)
		if err != nil {
			return err
		}
		if len(mods) == 0 {
			return nil
		}
		for _, mod := range mods {
			if mod.Timestamp.After(to) {
				return nil
			}
			mchan <- mod
		}
		from = mods[len(mods)-1].Timestamp
	}
}

func fpage(ctx context.Context, t time.Time) ([]modv, error) {
	url := fmt.Sprintf("https://index.golang.org/index/since=%s", t.String())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m := make([]modv, 0, 2000)
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func ndays(y, m int) int {
	leap := y%4 == 0 && (y%100 != 0 || y%400 == 0)
	d := days[m-1]
	if leap && m == 2 {
		d++
	}
	return d
}

var days = [...]int{
	31,
	28,
	31,
	30,
	31,
	30,
	31,
	31,
	30,
	31,
	30,
	31,
}
