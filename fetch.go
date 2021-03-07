package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	pageSize    = 2000
	bigPageSize = pageSize * 16
	fetchWindow = time.Hour * 24 * 30
	apiRetries  = 4
)

var origin = time.Date(2019, 04, 10, 19, 8, 52, 997264, time.UTC)

type fetcherParallel struct {
	cache bool
}

func (f fetcherParallel) fetch(ctx context.Context, ochan chan<- modv) error {
	if f.cache {
		_ = os.Mkdir("cache", os.ModePerm)
	}
	defer close(ochan)
	g, ctx := errgroup.WithContext(ctx)
	t := time.Now().UTC()
	for t.After(origin) {
		from, to := t.Add(-fetchWindow), t
		if from.Before(origin) {
			from = origin
		}
		g.Go(func() error {
			f := fetcherInterval{
				cache: f.cache,
				from:  from,
				to:    to,
			}
			return f.fetch(ctx, ochan)
		})
		t = from
	}
	return g.Wait()
}

type fetcherInterval struct {
	cache    bool
	from, to time.Time
}

func (f fetcherInterval) fetch(ctx context.Context, ochan chan<- modv) error {
	fname := fmt.Sprintf(
		"cache/page_%s_%s.json",
		f.from.Format(time.RFC3339Nano),
		f.to.Format(time.RFC3339Nano),
	)
	cache := make([]modv, 0, bigPageSize)
	if f.cache {
		if mods, err := fromFile(ctx, fname); err == nil {
			for _, mod := range mods {
				ochan <- mod
			}
			return nil
		}
	}
	defer func() {
		if f.cache {
			_ = toFile(ctx, fname, cache)
		}
	}()
	for {
		var mods []modv
		if err := retry(ctx, apiRetries, func(ctx context.Context) error {
			ms, err := fetchAPI(ctx, f.from)
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
		cache = append(cache, mods...)
		for _, mod := range mods {
			if mod.Timestamp.After(f.to) {
				return nil
			}
			ochan <- mod
		}
		f.from = mods[len(mods)-1].Timestamp
	}
}

func fromFile(ctx context.Context, fname string) ([]modv, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	mods := make([]modv, 0, bigPageSize)
	if err := json.Unmarshal(b, &mods); err != nil {
		return nil, err
	}
	return nil, f.Close()
}

func toFile(ctx context.Context, fname string, mods []modv) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	b, err := json.Marshal(mods)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		return err
	}
	return f.Close()
}

func fetchAPI(ctx context.Context, t time.Time) ([]modv, error) {
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
	mods := make([]modv, 0, pageSize)
	if err := json.Unmarshal(fixJson(b), &mods); err != nil {
		return nil, err
	}
	return mods, nil
}

func fixJson(b []byte) []byte {
	if len(b) == 0 {
		return []byte("[]")
	}
	sbuf := string(b)
	m := strings.Count(sbuf, "}")
	sbuf = strings.Replace(sbuf, "}", "},", m-1)
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
