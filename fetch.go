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
	psize    = 2000
	fwindow  = time.Hour * 24 * 30
	mretries = 4
)

type fetcher interface {
	fetch(ctx context.Context, ochan chan<- modv) error
}

type fcache struct {
	f fetcher
}

func (f fcache) fetch(ctx context.Context, ochan chan<- modv) error {
	defer close(ochan)
	if ff, err := os.Open("cache.json"); err == nil {
		b, err := ioutil.ReadAll(ff)
		if err != nil {
			return err
		}
		mods := make([]modv, 0, psize)
		if err := json.Unmarshal(b, &mods); err != nil {
			return err
		}
		for _, mod := range mods {
			ochan <- mod
		}
		return ff.Close()
	}
	ff, err := os.Create("cache.json")
	if err != nil {
		return err
	}
	if _, err := ff.Write([]byte("[")); err != nil {
		return err
	}
	g, ctx := errgroup.WithContext(ctx)
	ch := make(chan modv, len(ochan))
	g.Go(func() error {
		for mod := range ch {
			b, err := json.Marshal(mod)
			if err != nil {
				return err
			}
			if _, err := ff.Write(b); err != nil {
				return err
			}
			if _, err := ff.Write([]byte(",\n")); err != nil {
				return err
			}
			ochan <- mod
		}
		return nil
	})
	if err := f.f.fetch(ctx, ch); err != nil {
		return err
	}
	if err := g.Wait(); err != nil {
		return err
	}
	if _, err := ff.Seek(-2, os.SEEK_END); err != nil {
		return err
	}
	if _, err := ff.Write([]byte("]")); err != nil {
		return err
	}
	return ff.Close()
}

type fmulti struct{}

func (f fmulti) fetch(ctx context.Context, ochan chan<- modv) error {
	defer close(ochan)
	g, ctx := errgroup.WithContext(ctx)
	t := time.Now().UTC()
	for t.After(origin) {
		from, to := t.Add(-fwindow), t
		if from.Before(origin) {
			from = origin
		}
		g.Go(func() error {
			f := fAPI{from: from, to: to}
			return f.fetch(ctx, ochan)
		})
		t = from
	}
	return g.Wait()
}

type fAPI struct {
	from, to time.Time
}

func (f fAPI) fetch(ctx context.Context, ochan chan<- modv) error {
	for {
		var mods []modv
		if err := retry(ctx, mretries, func(ctx context.Context) error {
			ms, err := fpage(ctx, f.from)
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
			if mod.Timestamp.After(f.to) {
				return nil
			}
			ochan <- mod
		}
		f.from = mods[len(mods)-1].Timestamp
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
	if err := json.Unmarshal(fjson(b), &mods); err != nil {
		return nil, err
	}
	return mods, nil
}

func fjson(b []byte) []byte {
	if len(b) == 0 {
		return []byte("[]")
	}
	sbuf := string(b)
	m := strings.Count(sbuf, "}")
	sbuf = strings.Replace(sbuf, "}", "},", m-1)
	return []byte(fmt.Sprintf("[%s]", sbuf))
}

type action func(context.Context) error

func retry(ctx context.Context, max int, a action) (err error) {
	t := time.Second / time.Duration(max)
	for i := 0; i < max; i++ {
		if err = a(ctx); err == nil {
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
