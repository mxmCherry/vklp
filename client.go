package vklp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type skipper struct{}

var Skip skipper

const (
	DefaultScheme  = "https"
	DefaultWait    = "25"
	DefaultMode    = "0"
	DefaultVersion = "1"
)

type Client interface {
	Next() error
	Decode(...interface{}) error
	Stop() error
}

type Options struct {
	Server  string
	Key     string
	TS      string
	Wait    string
	Mode    string
	Version string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(options Options) (Client, error) {
	return From(new(http.Client), options)
}

func From(httpClient HTTPClient, options Options) (Client, error) {
	u, err := url.Parse(options.Server)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = DefaultScheme
	}

	if options.Wait == "" {
		options.Wait = DefaultWait
	}
	if options.Mode == "" {
		options.Mode = DefaultMode
	}
	if options.Version == "" {
		options.Version = DefaultVersion
	}

	q := u.Query()
	q.Set("act", "a_check")
	q.Set("key", options.Key)
	q.Set("ts", options.TS)
	q.Set("wait", options.Wait)
	q.Set("mode", options.Mode)
	q.Set("version", options.Version)
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithCancel(context.Background())

	r := bytes.NewReader(nil)
	d := json.NewDecoder(r)

	return &client{
		http:  httpClient,
		url:   u,
		query: q,

		ctx:    ctx,
		cancel: cancel,

		reader: r,
		json:   d,
	}, nil
}

// ----------------------------------------------------------------------------

const jsonArrayOpener = json.Delim('[')

type client struct {
	http  HTTPClient
	url   *url.URL
	query url.Values

	ctx    context.Context
	cancel context.CancelFunc

	reader *bytes.Reader
	json   *json.Decoder

	updates [][]byte
}

func (c *client) Stop() error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}

	c.cancel()
	return nil
}

func (c *client) Next() error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}

	if n := len(c.updates); n != 0 {
		upd := c.updates[0]
		if n > 1 {
			c.updates = c.updates[1:]
		} else {
			c.updates = nil
		}

		c.reader.Reset(upd)
		c.json = json.NewDecoder(c.reader)

		if t, err := c.json.Token(); err != nil {
			return err
		} else if delim, ok := t.(json.Delim); !ok || delim != jsonArrayOpener {
			return fmt.Errorf("vklp: expected JSON to start with array opening bracket '[', but got: %#v", t)
		}
		return nil
	}

	req, err := http.NewRequest("GET", c.url.String(), nil)
	if err != nil {
		return err
	}
	req = req.WithContext(c.ctx)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}

	wrapper := new(struct {
		TS      json.RawMessage   `json:"ts,omitempty"`
		Updates []json.RawMessage `json:"updates,omitempty"`
		Failed  uint8             `json:"failed,omitempty"`
	})
	if err = json.NewDecoder(resp.Body).Decode(wrapper); err != nil {
		return err
	}
	if wrapper.Failed != 0 {
		return Error(wrapper.Failed)
	}

	c.updates = make([][]byte, len(wrapper.Updates))
	for i, upd := range wrapper.Updates {
		c.updates[i] = upd
	}

	c.query.Set("ts", string(wrapper.TS))
	c.url.RawQuery = c.query.Encode()

	return c.Next()
}

func (c *client) Decode(vs ...interface{}) error {
	for i := 0; i < len(vs); i++ {
		if !c.json.More() {
			return nil
		}
		if vs[i] == Skip {
			var discard interface{}
			vs[i] = &discard
		}
		if err := c.json.Decode(vs[i]); err != nil {
			return err
		}
	}
	return nil
}
