# wswrapper
Wrapper of gorilla/websocket

## Example
```go
import (
	ws "github.com/dictor/wswrapper"
)
```
```go
hub := ws.NewHub()
go hub.Run(wsEvent)
```
```go
func wsEvent(evt *ws.WebsocketEvent) {
	switch evt.Kind {
	case ws.EVENT_RECIEVE:
    str := string(*evt.Msg)
	case ws.EVENT_REGISTER:
		
	case ws.EVENT_UNREGISTER:
		
	case ws.EVENT_ERROR:
		
	}
}
```

