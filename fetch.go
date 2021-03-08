package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	cacheDir    = ".cache"
)

var origin = time.Date(2019, 04, 10, 19, 8, 52, 997264, time.UTC)

func fetch(ctx context.Context, cache, verbose bool) <-chan modv {
	ch := make(chan modv, bigPageSize)
	f := fetcherParallel{cache, verbose}
	go func() {
		defer close(ch)
		if err := f.fetch(ctx, ch); verbose && err != nil {
			log.Println(err)
		}
	}()
	return ch
}

type fetcherParallel struct {
	cache, verbose bool
}

func (f fetcherParallel) fetch(ctx context.Context, ochan chan<- modv) error {
	if f.cache {
		if err := os.Mkdir(cacheDir, os.ModePerm); f.verbose && err != nil {
			log.Println(err)
		}
	}
	g, ctx := errgroup.WithContext(ctx)
	now := time.Now().UTC()
	for t := origin; t.Before(now); {
		cache, from, to := f.cache, t, t.Add(fetchWindow)
		if to.After(now) {
			to = now
			cache = false
		}
		g.Go(func() error {
			f := fetcherInterval{cache, f.verbose, from, to}
			return f.fetch(ctx, ochan)
		})
		t = to
	}
	return g.Wait()
}

type fetcherInterval struct {
	cache, verbose bool
	from, to       time.Time
}

func (f fetcherInterval) fetch(ctx context.Context, ochan chan<- modv) error {
	fname := fmt.Sprintf(
		"%s/page_%s_%s.json",
		cacheDir,
		f.from.Format(time.RFC3339Nano),
		f.to.Format(time.RFC3339Nano),
	)
	cache := make([]modv, 0, bigPageSize)
	if f.cache {
		mods, err := fromFile(ctx, fname)
		if err != nil {
			if f.verbose {
				log.Println(err)
			}
		} else {
			for _, mod := range mods {
				ochan <- mod
			}
			return nil
		}
	}
	defer func() {
		if f.cache {
			if err := toFile(ctx, fname, cache); f.verbose && err != nil {
				log.Println(err)
			}
		}
	}()
	for {
		var mods []modv
		if err := retry(ctx, f.verbose, apiRetries, func(ctx context.Context) error {
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
		for _, mod := range mods {
			if mod.Timestamp.After(f.to) {
				return nil
			}
			cache = append(cache, mod)
			ochan <- mod
		}
		f.from = mods[len(mods)-1].Timestamp.Add(time.Nanosecond)
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
	return mods, f.Close()
}

func toFile(ctx context.Context, fname string, mods []modv) error {
	hash := sha256.Sum256([]byte(fname))
	tmpname := hex.EncodeToString(hash[:])
	f, err := ioutil.TempFile("", tmpname)
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
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), fname)
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

func retry(ctx context.Context, verbose bool, max int, f func(context.Context) error) (err error) {
	t := time.Second / time.Duration(max)
	for i := 0; i < max; i++ {
		if err = f(ctx); err == nil {
			if verbose {
				fmt.Println(err)
			}
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
