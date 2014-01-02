package server

import (
	"github.com/vmihailenco/msgpack"
)

const (
	PLAYER_LEFT       = iota
	PLAYER_DISCONNECT = iota

	GROUP_ALL = iota

	MESSAGE_INVALID = 0
	MESSAGE_CHAT    = 1
)

type Msg struct {
	MessageType int `msgpack:"_t"`
}

type ChatMsg struct {
	Msg
	Name    string `msgpack:"n"`
	Message string `msgpack:"m"`
	Server  bool   `msgpack:"s"`
}

func ChatMessage(from *Player, message string) []byte {
	var c ChatMsg
	c.MessageType = MESSAGE_CHAT
	if from == nil {
		c.Server = true
	} else {
		c.Server = false
		c.Name = from.User.Name
	}
	c.Message = message

	b, _ := msgpack.Marshal(c)
	return b
}

func MessageType(message []byte) int {
	var mt Msg
	mt.MessageType = MESSAGE_INVALID
	msgpack.Unmarshal(message, &mt)
	return mt.MessageType
}
