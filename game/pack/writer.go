package pack

import (
	"bytes"
	"encoding/binary"
	"errors"
	t "game/typedefine"
	jsoniter "github.com/json-iterator/go"
	"reflect"
)

var (
	writerErr = errors.New("writer error")
)

func init() {
	t.EncodeWriter = encodeWriter
	t.PackWriter = packWriter
}

type Writer bytes.Buffer

func NewPack(sys, cmd int16, data ...interface{}) *Writer {
	writer := &Writer{}
	writer.Writer(sys, cmd)
	writer.Writer(data...)
	return writer
}

func NewWriter(data ...interface{}) *Writer {
	writer := &Writer{}
	writer.Writer(data...)
	return writer
}

func GetBytes(data ...interface{}) []byte {
	writer := &Writer{}
	writer.Writer(data...)
	return writer.Bytes()
}

func (this *Writer) Writer(datas ...interface{}) {
	writer := this.parse()
	for _, data := range datas {
		switch val := data.(type) {
		case int8, uint8, int16, uint16, int32, uint32, uint64, float64, int64:
			if err := binary.Write(writer, littleEndian, val); err != nil {
				panic(writerErr)
			}
		case int:
			var tv = int32(val)
			if err := binary.Write(writer, littleEndian, tv); err != nil {
				panic(writerErr)
			}
		case uint:
			var tv = uint32(val)
			if err := binary.Write(writer, littleEndian, tv); err != nil {
				panic(writerErr)
			}
		case []byte:
			writer.Write(val)
		case string:
			var l = uint16(len(val))
			if err := binary.Write(writer, littleEndian, l); err != nil {
				panic(writerErr)
			}
			writer.Write([]byte(val))
		case bool:
			var v byte
			if val {
				v = 1
			}
			if err := binary.Write(writer, littleEndian, v); err != nil {
				panic(writerErr)
			}
		case *t.CrossActor:
			buf, _ := jsoniter.Marshal(val)
			if err := binary.Write(writer, littleEndian, int32(len(buf))); err != nil {
				panic(writerErr)
			}
			writer.Write(buf)
		default:
			panic("writer invalid type " + reflect.TypeOf(data).String())
		}
	}
}

func (this *Writer) Len() int {
	return this.parse().Len()
}

func (this *Writer) Bytes() []byte {
	return this.parse().Bytes()
}

func (this *Writer) parse() *bytes.Buffer {
	return (*bytes.Buffer)(this)
}

func packWriter(sys, cmd int16, data ...interface{}) []byte {
	return NewPack(sys, cmd, data...).Bytes()
}

func encodeWriter(writer interface{}) []byte {
	return writer.(*Writer).Bytes()
}
