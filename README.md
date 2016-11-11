# vklp

Low-level Go (Golang) [vk.com longpoll](https://vk.com/dev/using_longpoll) client

WARNING - work in progress!


# TODO

- [x] review update user ID type - probably, can be negative (groups / public pages: `userID = groupID + 1_000_000_000`)
- [ ] review [vk.com/dev/using_longpoll](https://vk.com/dev/using_longpoll) - Russian version seems to be more accurate/detailed
- [ ] update pooling (`sync.Pool`, `.Release` method on each update type)
- [ ] tests (`Client`, `unmarshalUpdate`)
- [ ] godoc comments
- [ ] more sophisticated types with custom `unmarshalJSON` (timestamp - `time.Time`, `uint8` - `bool`, attachments - should be a list of `Attachment` etc)
- [x] think of possibility of returning `[]byte` to allow clients handle updates, unsupported by this lib (some events are not even documented on [vk.com/dev/using_longpoll](https://vk.com/dev/using_longpoll))
- [ ] think of possibility to store both update object and unmarshaller error (should never occur, but theoretically will allow to skip "broken" updates instead of rejecting entire batch)


# Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/mxmCherry/vkapi"
	"github.com/mxmCherry/vklp"
)

func main() {

	// get longpoll connection details first:

	api := vkapi.New(vkapi.Options{
		AccessToken: "YOUR_ACCESS_TOKEN",
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

	lp, err := vklp.New(vklp.Options{
		Server:  lpRes.Server,
		Key:     lpRes.Key,
		TS:      time.Unix(lpRes.TS, 0),
		Mode:    vklp.ModeAttachments | vklp.ModeExtendedEvents | vklp.ModePTS | vklp.ModeFriendOnlineExtra | vklp.ModeMessageRandomID,
		Version: vklp.V1,
	})
	if err != nil {
		panic(err.Error())
	}
	defer lp.Stop()

	for {
		update, err := lp.Next()
		if err != nil {
			panic(err.Error())
		}

		switch upd := update.(type) {

		case *vklp.AddNewMessage:
			fmt.Printf("new message from %d: %s\n", upd.FromID, upd.Text)

		case *vklp.FriendOnline:
			fmt.Printf("friend online: %d\n", upd.UserID)

		case *vklp.FriendOffline:
			fmt.Printf("friend offline: %d\n", upd.UserID)

		case []byte:
			// some events are not supported by this lib,
			// or even are not documented on https://vk.com/dev/using_longpoll ,
			// so these updates are returned as raw JSON byte slice:
			fmt.Printf("unknown update: %s\n", string(upd))

		default:
			// reject updates, that you are not interested in:
			fmt.Printf("rejected update: %#v\n", upd)

		} // end switch
	} // end for
}
```
