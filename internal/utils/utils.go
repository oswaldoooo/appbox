package utils

import (
	"errors"
	"io"
	"os"
	"sync"

	"github.com/emirpasic/gods/v2/maps/treemap"
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

type iosession struct {
	io.Writer
	cond *sync.Cond
	lock sync.Mutex
}

func (self *iosession) wait() {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.cond.Wait()
}
func (self *iosession) notify() {
	self.cond.Signal()
}

type IOBroadcastor struct {
	waiter *treemap.Map[string, *iosession]
	lock   sync.Mutex
}

func (self *IOBroadcastor) Write(b []byte) (n int, err error) {
	type kv struct {
		name    string
		element *iosession
	}
	var (
		size       int
		removelist []kv = make([]kv, 0, self.waiter.Size())
	)
	self.waiter.All(func(key string, value *iosession) bool {
		blen := len(b)
		offset := 0
	wsize:
		size, err = value.Write(b[offset:])
		if err != nil || size == 0 {
			removelist = append(removelist, kv{name: key, element: value})
			return true
		}
		if size < blen {
			offset += size
			goto wsize
		}
		return true
	})
	for _, k := range removelist {
		self.waiter.Remove(k.name)
		k.element.notify()
	}
	n = len(b)
	return
}
func (i *IOBroadcastor) Put(name string, w io.Writer) error {
	if w == nil {
		return errors.New("writer is nil")
	}
	i.lock.Lock()
	defer i.lock.Unlock()
	v := &iosession{
		Writer: w,
	}
	v.cond = sync.NewCond(&v.lock)
	i.waiter.Put(name, v)
	return nil
}

func (i *IOBroadcastor) Remove(name string) {
	target, ok := i.waiter.Get(name)
	if !ok {
		return
	}
	target.notify()
}

func (i *IOBroadcastor) Wait(name string) {
	target, ok := i.waiter.Get(name)
	if !ok {
		return
	}
	target.wait()
}

func (i *IOBroadcastor) Notify(name string) {
	target, ok := i.waiter.Get(name)
	if !ok {
		return
	}
	target.notify()
}
