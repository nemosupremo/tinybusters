package server

import (
	"fmt"
	"github.com/vmihailenco/msgpack"
)

type Lobby struct {
	*Room
}

func NewLobby() *Lobby {
	l := new(Lobby)
	l.Room = NewRoom(0)
	return l
}

func (l *Lobby) Register(p *Player) {
	l.Room.Register(p)
	p.AddReader(l)
	l.Broadcast(nil, ChatMessage(nil, fmt.Sprintf("Player %s joined the server.", p.Name)))
}

func (l *Lobby) Unregister(p *Player, group, code int) {
	l.Room.Unregister(p, group, code)
	p.RemoveReader(l)
	if group == GROUP_ALL || group == l.Group() {
		l.Broadcast(nil, ChatMessage(nil, fmt.Sprintf("Player %s left the server.", p.Name)))
	}
}

func (l *Lobby) Listen() {
	for message := range l.messages {
		switch MessageType(message.Message) {
		case MESSAGE_CHAT:
			var chat ChatMsg
			if err := msgpack.Unmarshal(message.Message, &chat); err == nil {
				l.Broadcast(nil, ChatMessage(message.Player, chat.Message))
			}
		}
	}
}
