package web

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

var (
	mh codec.MsgpackHandle
)

func init() {
	mh.WriteExt = true

}

func encode(d interface{}) []byte {
	var buf bytes.Buffer
	enc := codec.NewEncoder(&buf, &mh)
	enc.Encode(d)
	return buf.Bytes()
}

func decode(v interface{}, b []byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(b), &mh)
	dec.Decode(&v)
}
