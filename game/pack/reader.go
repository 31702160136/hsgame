package pack

import (
	"bytes"
	"encoding/binary"
	"errors"
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
				panic(readerErr)
			}
		case *int:
			var tv int32
			if err := binary.Read(reader, littleEndian, &tv); err != nil {
				panic(readerErr)
			}
			*val = int(tv)
		case *uint:
			var tv uint32
			if err := binary.Read(reader, littleEndian, &tv); err != nil {
				panic(readerErr)
			}
			*val = uint(tv)
		case []byte:
			if len(val) > 0 {
				if _, err := reader.Read(val); err != nil {
					panic(readerErr)
				}
			}
		case *string:
			var l uint16
			if err := binary.Read(reader, littleEndian, &l); err != nil {
				panic(readerErr)
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
				panic(readerErr)
			}
			if v == 1 {
				*val = true
			} else {
				*val = false
			}
		default:
			panic("reader invalid type " + reflect.TypeOf(data).String())
		}
	}
}

func (this *Reader) parse() *bytes.Reader {
	return (*bytes.Reader)(this)
}
