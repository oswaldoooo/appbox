package boxd

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

//this package is for box use

var (
	httport, tcport uint16
	Ip              net.IP
	//the parent process create will generate a random number,and child process use this number to exchange self pid from boxd
	Pid uint64
)

type Stream struct {
	conn   *net.TCPConn
	secret uint64
}

func NewStream(secret uint64, id uint16) (*Stream, error) {
	conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{
		IP:   Ip,
		Port: int(tcport),
	})
	if err != nil {
		return nil, err
	}
	err = handle_conn(secret, id, conn)
	if err != nil {
		return nil, errors.New("handle error " + err.Error())
	}
	var stream = Stream{
		conn:   conn,
		secret: secret,
	}
	return &stream, nil
}
func (s *Stream) Write(b []byte) (n int, err error) {
	return s.Write(b)
}
func handle_conn(secret uint64, id uint16, rw io.ReadWriter) error {
	var rd [10]byte
	binary.BigEndian.PutUint64(rd[:8], secret)
	binary.BigEndian.PutUint16(rd[8:10], id)
	size, err := rw.Write(rd[:])
	if err != nil || size == 0 {
		if err == nil {
			err = io.EOF
		}
		return err
	}
	err = read_util(rw, rd[:8])
	if err != nil {
		return err
	}
	//get real pid from boxd
	Pid = binary.BigEndian.Uint64(rd[:8])
	return nil
}

func read_util(r io.Reader, b []byte) error {
	blen := len(b)
	offset := 0
rdmsg:
	size, err := r.Read(b[offset:])
	if err != nil || size == 0 {
		if err == nil {
			err = io.EOF
		}
		return err
	}
	offset += size
	if offset < blen {
		goto rdmsg
	}
	return nil
}
