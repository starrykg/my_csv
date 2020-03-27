package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var PageEnd = errors.New("page END")

func main() {
	a := RVcs{File: "example.csv"}

	readF, err := a.ReadCvs()
	if err != nil {
		fmt.Println("error is", err)
	}
	for key, value := range readF.(map[int]map[int]string) {
		fmt.Printf(" %v,%v\n", key, value)
	}
}

type RVcs struct {
	Index int
	Len   int
	Buf   string
	File  string
}

func (r *RVcs) ReadCvs() (result interface{}, err error) {
	defer func() {
		if x := recover(); x != nil {
			err = errors.New("ReadCvs error")
			return
		}
	}()
	dat, err := r.ReadFile(r.File)
	if err != nil {
		return nil, err
	}
	readStr := make(map[int]map[int]string)
	r.Index = 0
	r.Buf = string(dat[:])
	r.Len = len(r.Buf) - 3
	var i int
	var j int
	for {
		str, err := r.readRecord()
		if readStr[i] == nil {
			readStr[i] = make(map[int]string)
		}
		readStr[i][j] = strings.Replace(str, "\n", "", -1)
		if err == PageEnd {
			break
		} else if err == io.EOF {
			i += 1
			j = 0
		}
		j += 1
	}
	return readStr, nil
}

func (r *RVcs) readRecord() (string, error) {
	var buf string
	var err error
	for {
		r.Index += 1
		if r.Len <= r.Index {
			err = PageEnd
			break
		}
		var countStr string
		countStr = r.Buf[r.Index-1 : r.Index]
		if strings.Count(countStr, "\r") == 1 {
			r.Index += 1
			countStr = r.Buf[r.Index-1 : r.Index]
			if strings.Count(countStr, "\n") == 1 {
				err = io.EOF
				break
			}
		} else if strings.Count(countStr, ",") == 1 {
			break
		}
		buf += countStr
	}
	return buf, err
}

func (r *RVcs) ReadFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var n int64 = 512

	if fi, err := f.Stat(); err == nil {
		if size := fi.Size() + 512; size > n {
			n = size
		}
	}
	return r.readAll(f, n)
}

func (r *RVcs) readAll(i io.Reader, capacity int64) (b []byte, err error) {
	var buf bytes.Buffer
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	if int64(int(capacity)) == capacity {
		buf.Grow(int(capacity))
	}
	_, err = buf.ReadFrom(i)
	return buf.Bytes(), err
}
