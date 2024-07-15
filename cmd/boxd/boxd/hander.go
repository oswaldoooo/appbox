package boxd

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/oswaldoooo/app/internal/utils"
)

type HandService struct {
	bind string
	websocket.Upgrader
	*StreamService
}

func (hs *HandService) connect2io(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	fd := query.Get("fd")
	pid := query.Get("pid")
	c, err := r.Cookie("session")
	if err != nil || len(c.Value) == 0 {
		return
	}
	conn, err := hs.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	sessmap, ok := hs.StreamService.sessionmap.Get(pid + "/" + fd)
	if !ok {
		return
	}
	brd, ok := sessmap.tmpdst.(*utils.IOBroadcastor)
	if !ok {
		return
	}
	brd.Put(c.Value, (*writer)(conn))
	brd.Wait(c.Value)
}

type writer websocket.Conn

func (wr *writer) Write(b []byte) (n int, err error) {
	wc := (*websocket.Conn)(wr)
	w, err := wc.NextWriter(websocket.TextMessage)
	if err != nil {
		return 0, err
	}
	return w.Write(b)
}
