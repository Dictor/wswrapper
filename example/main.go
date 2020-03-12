package main // import "github.com/dictor/wswrapper/example"

import (
	"fmt"
	ws "github.com/dictor/wswrapper"
	"io/ioutil"
	"net/http"
)

func main() {
	hub := ws.NewHub()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.AddClient(w, r)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f, _ := ioutil.ReadFile("front.html")
		w.Header().Add("Content-Type", "text/html")
		w.Write(f)
	})
	go hub.Run(wsEvent)
	fmt.Println(http.ListenAndServe(":80", nil))
}

func wsEvent(evt *ws.WebsocketEvent) {
	switch evt.Kind {
	case ws.EVENT_RECIEVE:
		str := string(*evt.Msg)
		evt.Client.Hub().SendAll([]byte(fmt.Sprintf("%d : %s", evt.Client.Id(), str)))
	case ws.EVENT_REGISTER:
		evt.Client.Send([]byte(fmt.Sprintf("[Server] Hello %d!", evt.Client.Id())))
		fmt.Printf("[WS_REG]%s\n", makeWsPrefix(evt.Client))
	case ws.EVENT_UNREGISTER:
		fmt.Printf("[WS_UNREG]%s\n", makeWsPrefix(evt.Client))
	case ws.EVENT_ERROR:
		fmt.Printf("[WS_ERROR]%s %s\n", makeWsPrefix(evt.Client), evt.Err)
	}
}

func makeWsPrefix(cli *ws.WebsocketClient) string {
	if cli == nil {
		return "(?)"
	}
	return fmt.Sprintf("(%s)(%d)", cli.Connection().RemoteAddr(), cli.Id())
}
