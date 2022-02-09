package typedefine

import (
	"game/pack"
	"github.com/gorilla/websocket"
	"sync"
)

type Account struct {
	ActorId      int64
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

func (this *Account) ReplyWriter(writer *pack.Writer) {
	if this.IsClose() {
		return
	}
	this.WriterMsg(writer.Bytes())
}

func (this *Account) Reply(sys, cmd int16, data ...interface{}) {
	if this.IsClose() {
		return
	}
	writer := pack.NewWriter(sys, cmd)
	writer.Writer(data...)
	this.WriterMsg(writer.Bytes())
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
