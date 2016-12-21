package main

import "io"
import "time"
import "strconv"

//import "fmt"

type RandomObject struct {
	Key    string
	Size   int64
	Md5    int64
	offset int64
	done   bool
}

func NewRandomObject(key string, prefix string, size int64) (*RandomObject, error) {
	k := key
	if k == "" {
		t := time.Now()
		k = prefix + t.Format("20060102150405") + strconv.Itoa(t.Nanosecond())
	}

	obj := &RandomObject{
		Key:    k,
		Size:   size,
		Md5:    0,
		offset: 0,
		done:   false,
	}

	return obj, nil
}

func (o *RandomObject) Read(p []byte) (int, error) {
	if o.offset >= o.Size {
		return 0, io.EOF
	}

	l := len(p)
	if o.offset+int64(l) > o.Size {
		l = int(o.Size - o.offset)
	}

	for i := int(0); i < l; i++ {
		p[i] = byte(time.Now().UnixNano())
	}

	o.offset += int64(l)

	//fmt.Println("Read ", l)
	return l, nil
}

func (o *RandomObject) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		o.offset = offset
	case io.SeekCurrent:
		o.offset += offset
	case io.SeekEnd:
		o.offset = o.Size
	}

	if o.offset > o.Size {
		o.offset = o.Size
	}

	return o.offset, nil
}
