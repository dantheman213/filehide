package compression

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

func CompressBinary(dat []byte) []byte {
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(dat)
	if err != nil {
		panic(err)
	}

	err = w.Close()
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func DecompressBinary(dat []byte) []byte {
	r, err := gzip.NewReader(bytes.NewBuffer(dat))
	if err != nil {
		panic(err)
	}
	defer r.Close()

	payload, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}

	return payload
}