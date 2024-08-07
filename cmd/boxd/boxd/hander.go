package boxd

import (
	"errors"
	"log"
	"net/http"

	"github.com/emirpasic/gods/v2/maps/treemap"
	"github.com/gorilla/websocket"
	"github.com/oswaldoooo/app/internal/utils"
)

const (
	InvalidArgs = "invalid args"
)

type HandService struct {
	bind string
	websocket.Upgrader
	*StreamService
	pidstore *treemap.Map[string, string]
	logger   *log.Logger
}

const (
	PASS    = "pass"
	REFUSED = "refused"
)

type status_t struct {
	status string
	reason string
}

func NewHandService(streamsvc *StreamService, bind string, logger *log.Logger) *HandService {
	var hs = HandService{
		pidstore:      streamsvc.pidmap,
		StreamService: streamsvc,
		logger:        logger,
		bind:          bind,
	}
	return &hs
}
func (hs *HandService) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/attch", hs.connect2io)
	mux.HandleFunc("/pid/put", hs.store_pid)

	return http.ListenAndServe(hs.bind, mux)
}
func (hs *HandService) connect2io(w http.ResponseWriter, r *http.Request) {
	var st status_t
	defer func() {
		hs.logger.Printf("%-10s %8s %20s\n", st.status, r.Method, st.reason, r.URL.RawPath, r.URL.RawQuery)
	}()
	query := r.URL.Query()
	fd := query.Get("fd")
	pid := query.Get("pid")
	c, err := r.Cookie("session")
	if err != nil || len(c.Value) == 0 {
		if err == nil {
			err = errors.New("not set pid session")
		}
		st.status = REFUSED
		st.reason = err.Error()
		return
	}
	conn, err := hs.Upgrade(w, r, nil)
	if err != nil {
		st.status = REFUSED
		st.reason = "websocket upgrade failed " + err.Error()
		return
	}
	sessmap, ok := hs.StreamService.sessionmap.Get(pid + "/" + fd)
	if !ok {
		st.status = REFUSED
		st.reason = "not found fd " + pid + "/" + fd
		return
	}
	brd, ok := sessmap.tmpdst.(*utils.IOBroadcastor)
	if !ok {
		st.status = REFUSED
		st.reason = "target not support io broadcast"
		return
	}
	st.status = PASS
	brd.Put(c.Value, (*writer)(conn))
	brd.Wait(c.Value)
}

func (hs *HandService) store_pid(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pid := query.Get("pid")
	secret := query.Get("secret")
	if len(secret) == 0 || len(pid) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(InvalidArgs))
		return
	}
	hs.pidstore.Put(secret, pid)
	w.WriteHeader(http.StatusOK)
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
