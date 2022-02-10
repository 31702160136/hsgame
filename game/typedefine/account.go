package typedefine

import (
	"github.com/gorilla/websocket"
	"sync"
)

var (
	EncodeWriter func(writer interface{}) []byte
	PackWriter   func(sys, cmd int16, data ...interface{}) []byte
)

type Account struct {
	Actor        *Actor
	AccountId    string
	Conn         *websocket.Conn
	IP           string
	data         [][]byte
	dataMux      *sync.Mutex
	readDataWait chan byte
	isWait       bool
	close        bool
}

func NewAccount(conn *websocket.Conn, ip string) *Account {
	account := &Account{
		Conn:         conn,
		IP:           ip,
		dataMux:      &sync.Mutex{},
		readDataWait: make(chan byte),
		data:         make([][]byte, 0),
	}
	return account
}

func (this *Account) ReadMsg() [][]byte {
	this.dataMux.Lock()
	data := this.data[:]
	this.data = this.data[len(this.data):]
	if len(data) == 0 {
		this.isWait = true
		this.dataMux.Unlock()
		<-this.readDataWait
	} else {
		this.dataMux.Unlock()
	}
	return data
}

func (this *Account) WriterMsg(data []byte) {
	this.dataMux.Lock()
	this.data = append(this.data, data)
	if this.isWait {
		this.isWait = false
		this.readDataWait <- 1
	}
	this.dataMux.Unlock()
}

func (this *Account) ReplyWriter(writer interface{}) {
	if this.IsClose() {
		return
	}
	this.WriterMsg(EncodeWriter(writer))
}

func (this *Account) Reply(sys, cmd int16, data ...interface{}) {
	if this.IsClose() {
		return
	}
	this.WriterMsg(PackWriter(sys, cmd, data...))
}

func (this *Account) SyncReply(msg []byte) {
	this.Conn.WriteMessage(websocket.BinaryMessage, msg)
}

func (this *Account) IsClose() bool {
	return this.close
}

func (this *Account) Close() {
	if this.close {
		return
	}
	this.dataMux.Lock()
	this.close = true
	if this.isWait {
		this.isWait = false
		close(this.readDataWait)
	}
	this.Conn.Close()
	this.data = nil
	this.dataMux.Unlock()
	this.dataMux = nil
}
