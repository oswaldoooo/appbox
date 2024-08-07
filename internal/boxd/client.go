package boxd

import (
	"errors"
	"io"
	"net/http"
	"strconv"
)

var client *http.Client
var __host string

func Init(addr string, _httport, _tcport uint16) {
	client = &http.Client{}
	__host = addr
	httport = _httport
	tcport = _tcport
}

func PutPid(secret, pid uint64) error {
	resp, err := client.Get("http://" + __host + ":" + strconv.Itoa(int(httport)) + "/pid/put?secret=" + strconv.FormatUint(secret, 10) + "&pid=" + strconv.FormatUint(pid, 10))
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	content, _ := io.ReadAll(resp.Body)
	return errors.New(string(content))
}
