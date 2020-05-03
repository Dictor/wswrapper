# wswrapper
**Wrapper of [gorilla/websocket](https://github.com/gorilla/websocket).** You don't need [Yak shaving](https://en.wiktionary.org/wiki/yak_shaving) with this library for creating websocket server based on gorilla/websocket.
[Document](https://pkg.go.dev/github.com/dictor/wswrapper?tab=doc), [Example](example)

## How to use
Import required packages.
```go
import (
	"fmt"
	ws "github.com/dictor/wswrapper"
	"io/ioutil"
	"net/http"
)
```

Implement websocket event handler function.
```go
func wsEvent(evt *ws.WebsocketEvent) {
	switch evt.Kind {
	case ws.EVENT_RECIEVE:
		//when some message received.
	case ws.EVENT_REGISTER:
		//when some client registered (connected).
	case ws.EVENT_UNREGISTER:
		//when some client unregisterd (disconnected).
	case ws.EVENT_ERROR:
		//when some error cause
	}
}
```

Register websocket server to web handler.
```go
func main() {
  //initiating new websocket server 
	hub := ws.NewHub()
  
  //endpoint for registering
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.AddClient(w, r)
	})
  
  //endpoint for frontend
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f, _ := ioutil.ReadFile("front.html")
		w.Header().Add("Content-Type", "text/html")
		w.Write(f)
	})
  
  //assign event handler and start websocket server (listening)
	go hub.Run(wsEvent)
  
  //start web server
	fmt.Println(http.ListenAndServe(":80", nil))
}
```
