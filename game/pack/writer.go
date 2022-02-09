package pack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"reflect"
)

var (
	writerErr = errors.New("writer error")
)

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

func (this *Writer) Writer(datas ...interface{}) {
	writer := this.parse()
	for _, data := range datas {
		switch val := data.(type) {
		case int8, uint8, int16, uint16, int32, uint32, uint64, float64, int64:
			if err := binary.Write(writer, littleEndian, val); err != nil {
				panic(readerErr)
			}
		case int:
			var tv int32
			if err := binary.Write(writer, littleEndian, tv); err != nil {
				panic(readerErr)
			}
		case uint:
			var tv uint32
			if err := binary.Write(writer, littleEndian, tv); err != nil {
				panic(readerErr)
			}
		case []byte:
			writer.Write(val)
		case string:
			var l = uint16(len(val))
			if err := binary.Write(writer, littleEndian, l); err != nil {
				panic(readerErr)
			}
			writer.Write([]byte(val))
		case bool:
			var v byte
			if val {
				v = 1
			}
			if err := binary.Write(writer, littleEndian, v); err != nil {
				panic(readerErr)
			}
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
