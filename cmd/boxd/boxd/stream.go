package boxd

import (
	"encoding/binary"
	"io"
	"net"
	"os"
	"path"
	"strconv"

	"github.com/emirpasic/gods/v2/maps/treemap"
)

type StreamSession struct {
	tmpdst io.Writer
}
type StreamService struct {
	stream_bind *net.TCPAddr
	sessionmap  *treemap.Map[string, *StreamSession]
}

func NewStreamService(bind string) (*StreamService, error) {
	addr, err := net.ResolveTCPAddr("tcp4", bind)
	if err != nil {
		return nil, err
	}
	var ss = StreamService{
		stream_bind: addr,
		sessionmap:  treemap.New[string, *StreamSession](),
	}
	return &ss, nil
}
func (ss *StreamService) Run() error {
	l, err := net.ListenTCP("tcp4", ss.stream_bind)
	if err != nil {
		return err
	}
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			break
		}
		go ss.handlecon(conn)
	}
	return err
}
func (ss *StreamService) handlecon(conn *net.TCPConn) {
	defer conn.Close()
	var buff [10]byte
	size, _ := conn.Read(buff[:])
	if size != 8 {
		return
	}
	boxpid := binary.BigEndian.Uint64(buff[:8])
	ffd := binary.BigEndian.Uint16(buff[8:])
	rpath := "/var/log/appbox/appbox" + strconv.FormatUint(boxpid, 10)
	os.MkdirAll(rpath, 0755)
	f, err := os.OpenFile(path.Join(rpath, strconv.Itoa(int(ffd))), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	session := &StreamSession{}
	ss.sessionmap.Put(strconv.FormatUint(boxpid, 10)+"/"+strconv.Itoa(int(ffd)), session)
	session.Copy(f, conn)
}

func (ss *StreamSession) Copy(dst io.Writer, src io.Reader) (size int, err error) {
	var (
		buff    [1 << 10]byte
		n, tmpn int
	)
	for {
		n, err = src.Read(buff[:])
		size += n
		if err != nil {
			return
		}
		if ss.tmpdst != nil {
			ss.tmpdst.Write(buff[:n])
		}
	sendmsg:
		tmpn, err = dst.Write(buff[:n])
		if err != nil {
			return
		}
		if tmpn < n {
			n -= tmpn
			goto sendmsg
		}
	}
}
