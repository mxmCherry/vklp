# vklp

Low-level Go (Golang) [vk.com longpoll](https://vk.com/dev/using_longpoll) client

WARNING - work in progress!


# Example

```go
package main

import (
	"fmt"

	"github.com/mxmCherry/vkapi"
	"github.com/mxmCherry/vklp"
)

func main() {

	// get longpoll connection details first:

	api := vkapi.New(vkapi.Options{
		AccessToken: "YOUR_ACCESS_TOKEN",
		Version:     "5.62",
	})

	lpReq := struct {
		UseSSL  bool `json:"use_ssl,omitempty"`
		NeedPTS bool `json:"need_pts,omitempty"`
	}{
		UseSSL: true,
	}

	lpRes := new(struct {
		Server string `json:"server"`
		Key    string `json:"key"`
		TS     int64  `json:"ts"`
	})

	err := api.Exec("messages.getLongPollServer", vkapi.ToParams(lpReq), lpRes)
	if err != nil {
		panic(err.Error())
	}

	// consume updates with longpoll:

	// instantiate longpoll client:
	lp, err := vklp.New(vklp.Options{
		Server:  lpRes.Server,
		Key:     lpRes.Key,
		TS:      lpRes.TS,
		Mode:    2 | 8 | 32 | 64 | 128,
		Version: "1",
	})
	if err != nil {
		panic(err.Error())
	}
	defer lp.Stop()

	// process events:
	for {

		// load next event:
		if err = lp.Next(); err != nil {
			panic(err.Error())
		}

		// decode event type:
		var t uint8
		if err := lp.Decode(&t); err != nil {
			panic(err.Error())
		}

		// decode events by their types:
		switch t {

		case 4: // received new message:
			var (
				messageID uint64
				peerID    int64
				subject   string
				text      string
			)
			// [$message_id, $flags, $peer_id, $timestamp, $subject, $text, $attachments, $random_id]
			if err = lp.Decode(&messageID, vklp.Skip, &peerID, vklp.Skip, &subject, &text); err != nil {
				panic(err.Error())
			}
			fmt.Printf("received new message %d from %d: %s - %s\n", messageID, peerID, subject, text)

		default: // drop unused events:
			fmt.Printf("dropped unused event %d\n", t)
		}

	} // end for
} // end main
```
