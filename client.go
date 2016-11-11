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
	ModeAttachments       uint8 = 2   // 2 — получать вложения
	ModeExtendedEvents    uint8 = 8   // 8 — возвращать расширенный набор событий
	ModePTS               uint8 = 32  // 32 — возвращать pts (это требуется для работы метода messages.getLongPollHistory без ограничения в 256 последних событий)
	ModeFriendOnlineExtra uint8 = 64  // 64 — в событии с кодом 8 (друг стал онлайн) возвращать дополнительные данные в поле $extra
	ModeMessageRandomID   uint8 = 128 // 128 — возвращать с сообщением параметр random_id (random_id может быть передан при отправке сообщения методом https://vk.com/dev/messages.send)
)

const (
	V0 uint8 = 0 // Для версии 0 (по умолчанию) идентификаторы сообществ будут приходить в формате group_id + 1000000000 для сохранения обратной совместимости
	V1 uint8 = 1 // Актуальная версия: 1
)

const (
	DefaultScheme = "https"
	DefaultWait   = 25 * time.Second
)

type Client interface {
	Next() (interface{}, error)
	Stop() error
}

type Options struct {
	Server  string
	Key     string
	TS      time.Time
	Wait    time.Duration
	Mode    uint8
	Version uint8
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
	q.Set("mode", strconv.FormatUint(uint64(options.Mode), 10))
	q.Set("version", strconv.FormatUint(uint64(options.Version), 10))

	wait := DefaultWait
	if options.Wait != 0 {
		wait = options.Wait
	}
	q.Set("wait", strconv.FormatInt(int64(wait.Seconds()), 10))

	u.RawQuery = q.Encode()

	ctx, cancel := context.WithCancel(context.Background())

	return &client{
		http:   httpClient,
		url:    u,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// ----------------------------------------------------------------------------

type client struct {
	http HTTPClient
	url  *url.URL

	ctx    context.Context
	cancel context.CancelFunc

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
		return unmarshalUpdate(upd)
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
		Failed  uint8             `json:"failed,omitempty"`
	})

	if err = json.NewDecoder(resp.Body).Decode(wrapper); err != nil {
		return nil, err
	}

	if wrapper.Failed != 0 {
		return nil, Error(wrapper.Failed)
	}

	c.updates = make([][]byte, len(wrapper.Updates))
	for i, upd := range wrapper.Updates {
		c.updates[i] = upd
	}

	q := c.url.Query()
	q.Set("ts", strconv.FormatInt(wrapper.TS, 10))
	c.url.RawQuery = q.Encode()

	return c.Next()
}
