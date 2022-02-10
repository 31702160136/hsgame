package pack

import (
	"bytes"
	t "cross/typedefine"
	"encoding/binary"
	"errors"
	jsoniter "github.com/json-iterator/go"
	"reflect"
)

var (
	readerErr = errors.New("read error")
)

type Reader bytes.Reader

func NewReader(data []byte) *Reader {
	return (*Reader)(bytes.NewReader(data))
}

func (this *Reader) Reset(data []byte) {
	reader := this.parse()
	reader.Reset(data)
}

func (this *Reader) Len() int {
	reader := this.parse()
	return reader.Len()
}

func (this *Reader) Read(datas ...interface{}) {
	reader := this.parse()
	for _, data := range datas {
		switch val := data.(type) {
		case *int8, *uint8, *int16, *uint16, *int32, *uint32, *int64, *uint64, *float64:
			if err := binary.Read(reader, littleEndian, val); err != nil {
				panic(err.Error())
			}
		case *int:
			var tv int32
			if err := binary.Read(reader, littleEndian, &tv); err != nil {
				panic(err.Error())
			}
			*val = int(tv)
		case *uint:
			var tv uint32
			if err := binary.Read(reader, littleEndian, &tv); err != nil {
				panic(err.Error())
			}
			*val = uint(tv)
		case []byte:
			if len(val) > 0 {
				if _, err := reader.Read(val); err != nil {
					panic(err.Error())
				}
			}
		case *string:
			var l uint16
			if err := binary.Read(reader, littleEndian, &l); err != nil {
				panic(err.Error())
			}
			s := make([]byte, l)
			n, _ := reader.Read(s)
			if uint16(n) < l {
				panic(readerErr)
			}
			*val = string(s)
		case *bool:
			var v byte
			if err := binary.Read(reader, littleEndian, &v); err != nil {
				panic(err.Error())
			}
			if v == 1 {
				*val = true
			} else {
				*val = false
			}
		case *t.CrossActor:
			var length int32
			if err := binary.Read(reader, littleEndian, &length); err != nil {
				panic(err.Error())
			}
			buf := make([]byte, length)
			if _, err := reader.Read(buf); err != nil {
				panic(err.Error())
			}
			if err := jsoniter.Unmarshal(buf, val); err != nil {
				panic(err.Error())
			}
		default:
			panic("reader invalid type " + reflect.TypeOf(data).String())
		}
	}
}

func (this *Reader) parse() *bytes.Reader {
	return (*bytes.Reader)(this)
}
