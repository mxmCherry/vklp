package vklp

import (
	"bytes"
	"sync"
)

var readerPool = newReaderPooler()

// ----------------------------------------------------------------------------

type readerPooler struct {
	pool sync.Pool
}

func newReaderPooler() readerPooler {
	return readerPooler{
		pool: sync.Pool{},
	}
}

func (p *readerPooler) For(b []byte) *bytes.Reader {
	x := p.pool.Get()
	if x == nil {
		return bytes.NewReader(b)
	}
	r := x.(*bytes.Reader)
	r.Reset(b)
	return r
}

func (p *readerPooler) Put(r *bytes.Reader) {
	p.pool.Put(r)
}
