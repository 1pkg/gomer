package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func fetch(ctx context.Context, from, to time.Time, mchan chan<- modv) error {
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
	url := fmt.Sprintf("https://index.golang.org/index?since=%s", t.Format(time.RFC3339Nano))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
	fmt.Println(string(b), url)
	m := make([]modv, 0, psize)
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}
