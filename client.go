package vklp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

var Skip struct{}

const (
	ModeAttachments = 2
	ModeExtended    = 8
	ModePTS         = 32
	ModeExtra       = 64
	ModeRandomID    = 128
)

type Client struct {
	http  HTTPClient
	url   *url.URL
	query url.Values

	ctx    context.Context
	cancel context.CancelFunc

	reader *bytes.Reader
	json   *json.Decoder

	updates [][]byte
}

type Options struct {
	Server  string
	Key     string
	TS      int64
	Wait    int64
	Mode    uint64
	Version string
}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

func New(options Options) (*Client, error) {
	return From(http.DefaultClient, options)
}

func From(httpClient HTTPClient, options Options) (*Client, error) {
	u, err := url.Parse(options.Server)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "https"
	}

	q := u.Query()
	q.Set("act", "a_check")
	q.Set("key", options.Key)
	q.Set("ts", strconv.FormatInt(options.TS, 10))
	q.Set("wait", strconv.FormatInt(options.Wait, 10))
	q.Set("mode", strconv.FormatUint(options.Mode, 10))
	q.Set("version", options.Version)
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithCancel(context.Background())

	r := bytes.NewReader(nil)
	d := json.NewDecoder(r)

	return &Client{
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

func (c *Client) Stop() error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	default:
	}

	c.cancel()
	return nil
}

func (c *Client) Next() error {
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

func (c *Client) Decode(vs ...interface{}) error {
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
