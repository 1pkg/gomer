package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func fetch(ctx context.Context, from, to time.Time, mchan chan<- modv) error {
	for {
		var mods []modv
		if err := retry(ctx, mretries, func(ctx context.Context) error {
			ms, err := fpage(ctx, from)
			if err != nil {
				return err
			}
			mods = ms
			return nil
		}); err != nil {
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
	mods := make([]modv, 0, psize)
	if err := json.Unmarshal(tojson(b), &mods); err != nil {
		return nil, err
	}
	return mods, nil
}

func tojson(b []byte) []byte {
	if len(b) == 0 {
		return []byte("[]")
	}
	sbuf := string(b)
	m := strings.Count(sbuf, "}")
	sbuf = strings.Replace(string(b), "}", "},", m-1)
	return []byte(fmt.Sprintf("[%s]", sbuf))
}

func retry(ctx context.Context, max int, f func(context.Context) error) (err error) {
	t := time.Second / time.Duration(max)
	for i := 0; i < max; i++ {
		if err = f(ctx); err == nil {
			return
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(t)
			t *= 2
		}
	}
	return
}
