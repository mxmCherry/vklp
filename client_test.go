package vklp_test

import (
	"net/http"

	"github.com/mxmCherry/vklp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var httpClient *mockHTTPClient
	var subject *vklp.Client

	BeforeEach(func() {
		httpClient = &mockHTTPClient{
			resp: &http.Response{
				Body: newMockReadCloser(`{
					"failed": 404
				}`),
			},
		}

		var err error
		subject, err = vklp.From(httpClient, vklp.Options{
			Server:  "some.domain/some/path?one=1&two=2",
			Key:     "DUMMY_KEY",
			TS:      1111111111,
			Wait:    1111,
			Mode:    111,
			Version: "11",
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should consume events", func() {
		type Message struct {
			Type        uint8
			MessageID   uint64
			Mask        uint64
			PeerID      int64
			Timestamp   int64
			Subject     string
			Text        string
			Attachments map[string]interface{}
			RandomID    int64
		}

		httpClient.resp.Body = newMockReadCloser(`{
			"ts": 2222222222,
			"updates": [
				[4,654321,8241,2000000202,1464950873,"My chat","Hello"],
				[4,1619489,561,123456,1464958914," ... ","hello", {"attach1_type":"photo","attach1":"123456_414233177", "attach2_type":"audio","attach2":"123456_456239018"}]
			]
		}`)

		msg := Message{}
		Expect(subject.Next()).To(Succeed())
		Expect(subject.Decode(
			&msg.Type,
			&msg.MessageID,
			&msg.Mask,
			&msg.PeerID,
			&msg.Timestamp,
			&msg.Subject,
			&msg.Text,
			&msg.Attachments,
			&msg.RandomID,
		)).To(Succeed())
		Expect(msg).To(Equal(Message{
			// [4,654321,8241,2000000202,1464950873,"My chat","Hello"],
			Type:        4,
			MessageID:   654321,
			Mask:        8241,
			PeerID:      2000000202,
			Timestamp:   1464950873,
			Subject:     "My chat",
			Text:        "Hello",
			Attachments: nil,
			RandomID:    0,
		}))

		Expect(httpClient.reqCount).To(Equal(1))
		Expect(httpClient.req).NotTo(BeNil())
		Expect(httpClient.req.URL.String()).To(Equal(
			"https://some.domain/some/path?act=a_check&key=DUMMY_KEY&mode=111&one=1&ts=1111111111&two=2&version=11&wait=1111",
		))

		msg = Message{}
		Expect(subject.Next()).To(Succeed())
		Expect(subject.Decode(
			&msg.Type,
			&msg.MessageID,
			&msg.Mask,
			&msg.PeerID,
			&msg.Timestamp,
			&msg.Subject,
			&msg.Text,
			&msg.Attachments,
			&msg.RandomID,
		)).To(Succeed())
		Expect(msg).To(Equal(Message{
			// [4,1619489,561,123456,1464958914," ... ","hello", {"attach1_type":"photo","attach1":"123456_414233177", "attach2_type":"audio","attach2":"123456_456239018"}]
			Type:      4,
			MessageID: 1619489,
			Mask:      561,
			PeerID:    123456,
			Timestamp: 1464958914,
			Subject:   " ... ",
			Text:      "hello",
			Attachments: map[string]interface{}{
				"attach1_type": "photo",
				"attach1":      "123456_414233177",
				"attach2_type": "audio",
				"attach2":      "123456_456239018",
			},
			RandomID: 0,
		}))

		Expect(httpClient.reqCount).To(Equal(1))

		httpClient.resp.Body = newMockReadCloser(`{
			"failed": 42
		}`)

		Expect(subject.Next()).To(MatchError("vklp: error 42"))
		Expect(httpClient.reqCount).To(Equal(2))
		Expect(httpClient.req).NotTo(BeNil())
		Expect(httpClient.req.URL.String()).To(Equal(
			"https://some.domain/some/path?act=a_check&key=DUMMY_KEY&mode=111&one=1&ts=2222222222&two=2&version=11&wait=1111",
		))
	})

})
