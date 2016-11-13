package vklp_test

import (
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVkLP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "vklp")
}

// ----------------------------------------------------------------------------

type mockHTTPClient struct {
	req      *http.Request
	reqCount int
	resp     *http.Response
	err      error
}

func (c *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.req = req
	c.reqCount++
	return c.resp, c.err
}

type mockReadCloser struct{ *strings.Reader }

func newMockReadCloser(s string) mockReadCloser {
	return mockReadCloser{
		Reader: strings.NewReader(s),
	}
}
func (rc mockReadCloser) Close() error {
	return nil
}
