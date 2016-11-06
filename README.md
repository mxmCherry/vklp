# vklp

Low-level Go (Golang) [vk.com longpoll](https://vk.com/dev/using_longpoll) client

WARNING - work in progress!


# Example

```go
package main

import (
	"log"
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
		Server: lpRes.Server,
		Key:    lpRes.Key,
		TS:     time.Unix(lpRes.TS, 0),
	})
	if err != nil {
		panic(err.Error())
	}
	defer lp.Stop()

	for {
		upd, err := lp.Next()
		if err != nil {
			panic(err.Error())
		}

		log.Println(string(upd)) // upd is raw JSON []byte slice
	}
}
```
