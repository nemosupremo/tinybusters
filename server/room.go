package server

import (
	"sync"
)

type Room struct {
	messages chan PlayerMessage
	players  map[*Player]bool
	group    int
	plock    *sync.RWMutex
}

func NewRoom(group int) *Room {
	return &Room{
		messages: make(chan PlayerMessage, 512),
		players:  make(map[*Player]bool),
		group:    group,
		plock:    new(sync.RWMutex),
	}
}

func (r *Room) Channel() chan<- PlayerMessage {
	return r.messages
}

func (r *Room) Register(p *Player) {
	r.plock.Lock()
	r.players[p] = true
	r.plock.Unlock()
}

func (r *Room) Unregister(p *Player, group, code int) {
	if group == GROUP_ALL || group == r.Group() {
		r.plock.Lock()
		delete(r.players, p)
		r.plock.Unlock()
	}
}

func (r *Room) Broadcast(except *Player, message []byte) {
	r.plock.RLock()
	for p, _ := range r.players {
		if p == except {
			continue
		}
		p.Write(message)
	}
	r.plock.RUnlock()
}

func (r *Room) Group() int {
	return r.group
}
