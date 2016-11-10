package vklp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	DefaultScheme = "https"
	DefaultMode   = ModeAttachmentsOff
	DefaultWait   = 25 * time.Second
)

type Client interface {
	Next() (interface{}, error)
	Stop() error
}

type Options struct {
	Server string
	Key    string
	TS     time.Time
	Mode   Mode
	Wait   time.Duration
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

	q := u.Query()
	q.Set("act", "a_check")
	q.Set("key", options.Key)
	q.Set("ts", strconv.FormatInt(options.TS.Unix(), 10))

	wait := DefaultWait
	if options.Wait != 0 {
		wait = options.Wait
	}
	q.Set("wait", strconv.FormatInt(int64(wait.Seconds()), 10))

	mode := DefaultMode
	if options.Mode != 0 {
		mode = options.Mode
	}
	q.Set("mode", strconv.FormatUint(uint64(mode), 10))

	u.RawQuery = q.Encode()

	ctx, cancel := context.WithCancel(context.Background())

	c := &client{
		http:   httpClient,
		url:    u,
		ctx:    ctx,
		cancel: cancel,
	}
	return c, nil
}

// ----------------------------------------------------------------------------

type client struct {
	http HTTPClient
	url  *url.URL

	ctx    context.Context
	cancel context.CancelFunc

	updates []interface{}
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

func (c *client) Next() (interface{}, error) {
	select {
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	default:
	}

	if n := len(c.updates); n != 0 {
		upd := c.updates[0]
		if n > 1 {
			c.updates = c.updates[1:]
		} else {
			c.updates = nil
		}
		return upd, nil
	}

	req, err := http.NewRequest("GET", c.url.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(c.ctx)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	wrapper := new(struct {
		TS      int64             `json:"ts,omitempty"`
		Updates []json.RawMessage `json:"updates,omitempty"`
		Failed  *Error            `json:"failed,omitempty"`
	})

	if err = json.NewDecoder(resp.Body).Decode(wrapper); err != nil {
		return nil, err
	}

	if wrapper.Failed != nil {
		return nil, *wrapper.Failed
	}

	updates := make([]interface{}, len(wrapper.Updates))
	for i, b := range wrapper.Updates {
		upd, err := unmarshalUpdate(b)
		if err != nil {
			return nil, err
		}
		updates[i] = upd
	}
	c.updates = updates

	q := c.url.Query()
	q.Set("ts", strconv.FormatInt(wrapper.TS, 10))
	c.url.RawQuery = q.Encode()

	return c.Next()
}
