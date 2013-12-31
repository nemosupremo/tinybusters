package server

import (
	"github.com/gorilla/websocket"
	"github.com/nemothekid/tinybusters/datastore"
	"sync"
	"time"
)

const (
	writeWait      = 3 * time.Second
	pongWait       = 5 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type PlayerReader interface {
	Channel() chan<- PlayerMessage
	Register(*Player)
	Unregister(*Player, int, int)
	Group() int
}

type Player struct {
	writech chan []byte
	readers []PlayerReader
	readerl *sync.RWMutex
	ws      *websocket.Conn
	User    *datastore.User
}

type PlayerMessage struct {
	Player  *Player
	Message []byte
}

func NewPlayer(ws *websocket.Conn, u *datastore.User) *Player {
	p := &Player{
		writech: make(chan []byte, 512),
		readers: make([]PlayerReader, 0, 64),
		ws:      ws,
		readerl: new(sync.RWMutex),
		User:    u,
	}
	return p
}

func (p *Player) Listen() {
	go p.writer()
	p.reader()
}

func (p *Player) Unregister(group, code int) {
	p.readerl.RLock()
	for _, reader := range p.readers {
		defer reader.Unregister(p, group, code)
	}
	p.readerl.RUnlock()
}

func (p *Player) reader() {
	defer func() {
		p.Unregister(GROUP_ALL, PLAYER_DISCONNECT)
		p.ws.Close()
	}()
	p.ws.SetReadDeadline(time.Now().Add(pongWait))
	p.ws.SetPongHandler(func(string) error { p.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := p.ws.ReadMessage()
		if err != nil {
			break
		}
		p.dispatchMessage(message)
	}
}

func (p *Player) AddReader(pl PlayerReader) {
	p.readerl.Lock()
	p.readers = append(p.readers, pl)
	p.readerl.Unlock()
}

func (p *Player) RemoveReader(pl PlayerReader) {
	p.readerl.Lock()
	for i := range p.readers {
		if p.readers[i] == pl {
			p.readers = append(p.readers[:i], p.readers[i+1:]...)
			break
		}
	}
	p.readerl.Unlock()
}

func (p *Player) dispatchMessage(message []byte) {
	p.readerl.RLock()
	pm := PlayerMessage{p, message}
	for _, reader := range p.readers {
		reader.Channel() <- pm
	}
	p.readerl.RUnlock()
}

func (p *Player) Write(message []byte) {
	select {
	case p.writech <- message:
	default:
		close(p.writech)
	}
}

func (p *Player) write(mt int, payload []byte) error {
	p.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return p.ws.WriteMessage(mt, payload)
}

func (p *Player) writer() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.ws.Close()
	}()
	for {
		select {
		case message, ok := <-p.writech:
			if !ok {
				p.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := p.write(websocket.BinaryMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := p.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
