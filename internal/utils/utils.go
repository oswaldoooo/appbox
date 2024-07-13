package utils

import (
	"errors"
	"io"
	"os"
)

func SliceConvert[T1, T2 any](dst []T1, src []T2, cf func(dst *T1, src T2)) []T1 {
	var (
		dstlen, srclen = len(dst), len(src)
	)
	if cap(dst) <= dstlen+srclen {
		newdst := make([]T1, dstlen, dstlen+srclen)
		copy(newdst, dst)
		dst = newdst
	}
	for i := range src {
		var t T1
		cf(&t, src[i])
		dst = append(dst, t)
	}
	return dst
}
func Copy(dst, src string) error {
	finfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(src, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	f2, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, finfo.Mode().Perm())
	if err != nil {
		return err
	}
	defer f2.Close()
	io.Copy(f2, f)
	return nil
}
func CopyAll(elem ...string) (err error) {
	if len(elem) <= 1 {
		err = errors.New("not set dst")
		return
	}
	dstpath := elem[len(elem)-1]
	var errlist = make([]error, len(elem)-1)
	for i := 0; i < len(elem)-1; i++ {
		errlist[i] = Copy(elem[i], dstpath)
	}
	err = errors.Join(errlist...)
	return
}
