package core

import (
	"strconv"
)

// Key represents one keyboard stroke.
type Key string

type keyEvent struct {
	keys  []Key
	event EventType
}

const (
	keysEq = iota
	keysPending
	keysNeq
)

func (ke keyEvent) cmp(ks []Key) int {
	if len(ke.keys) < len(ks) {
		return keysNeq
	}
	for i, k := range ke.keys {
		if i >= len(ks) {
			return keysPending
		}
		if k != ks[i] {
			return keysNeq
		}
	}
	return keysEq
}

// KeyManager holds the key mappings and current key sequence.
type KeyManager struct {
	keys      []Key
	keyEvents []keyEvent
	count     bool
}

// NewKeyManager creates a new KeyManager.
func NewKeyManager(count bool) *KeyManager {
	return &KeyManager{count: count}
}

// Register adds a new key mapping.
func (km *KeyManager) Register(event EventType, keys ...Key) {
	km.keyEvents = append(km.keyEvents, keyEvent{keys, event})
}

// Press checks the new key down event.
func (km *KeyManager) Press(k Key) Event {
	km.keys = append(km.keys, k)
	for i := 0; i < len(km.keys); i++ {
		keys := km.keys[i:]
		var count int64
		if km.count {
			numStr := ""
			for j, k := range keys {
				if len(k) == 1 && ('1' <= k[0] && k[0] <= '9' || k[0] == '0' && j > 0) {
					numStr += string(k)
				} else {
					break
				}
			}
			keys = keys[len(numStr):]
			count, _ = strconv.ParseInt(numStr, 10, 64)
		}
		for _, ke := range km.keyEvents {
			switch ke.cmp(keys) {
			case keysPending:
				return Event{Type: EventNop}
			case keysEq:
				km.keys = nil
				return Event{Type: ke.event, Count: count}
			}
		}
	}
	km.keys = nil
	return Event{Type: EventNop}
}

func defaultKeyManagers() map[Mode]*KeyManager {
	kms := make(map[Mode]*KeyManager)
	km := NewKeyManager(true)
	km.Register(EventQuit, "Z", "Q")
	km.Register(EventCursorUp, "up")
	km.Register(EventCursorDown, "down")
	km.Register(EventCursorLeft, "left")
	km.Register(EventCursorRight, "right")
	km.Register(EventPageUp, "pgup")
	km.Register(EventPageDown, "pgdn")
	km.Register(EventPageTop, "home")
	km.Register(EventPageEnd, "end")
	km.Register(EventCursorUp, "k")
	km.Register(EventCursorDown, "j")
	km.Register(EventCursorLeft, "h")
	km.Register(EventCursorRight, "l")
	km.Register(EventCursorPrev, "b")
	km.Register(EventCursorNext, "w")
	km.Register(EventCursorHead, "0")
	km.Register(EventCursorHead, "^")
	km.Register(EventCursorEnd, "$")
	km.Register(EventScrollUp, "c-y")
	km.Register(EventScrollDown, "c-e")
	km.Register(EventPageUp, "c-b")
	km.Register(EventPageDown, "c-f")
	km.Register(EventPageUpHalf, "c-u")
	km.Register(EventPageDownHalf, "c-d")
	km.Register(EventPageTop, "g", "g")
	km.Register(EventPageEnd, "G")
	km.Register(EventJumpTo, "c-]")
	km.Register(EventJumpBack, "c-t")
	km.Register(EventDeleteByte, "x")
	km.Register(EventDeletePrevByte, "X")
	km.Register(EventIncrement, "c-a")
	km.Register(EventIncrement, "+")
	km.Register(EventDecrement, "c-x")
	km.Register(EventDecrement, "-")

	km.Register(EventStartInsert, "i")
	km.Register(EventStartInsertHead, "I")
	km.Register(EventStartAppend, "a")
	km.Register(EventStartAppendEnd, "A")
	km.Register(EventStartReplaceByte, "r")
	km.Register(EventStartReplace, "R")

	km.Register(EventStartCmdline, ":")
	kms[ModeNormal] = km

	km = NewKeyManager(false)
	km.Register(EventExitInsert, "escape")
	km.Register(EventExitInsert, "c-c")
	km.Register(EventCursorUp, "up")
	km.Register(EventCursorDown, "down")
	km.Register(EventCursorLeft, "left")
	km.Register(EventCursorRight, "right")
	km.Register(EventPageUp, "pgup")
	km.Register(EventPageDown, "pgdn")
	km.Register(EventPageTop, "home")
	km.Register(EventPageEnd, "end")
	km.Register(EventBackspace, "backspace")
	km.Register(EventBackspace, "backspace2")
	km.Register(EventDelete, "delete")
	kms[ModeInsert] = km
	kms[ModeReplace] = km

	km = NewKeyManager(false)
	km.Register(EventSpaceCmdline, "space")
	km.Register(EventCursorLeft, "left")
	km.Register(EventCursorLeft, "c-b")
	km.Register(EventCursorRight, "right")
	km.Register(EventCursorRight, "c-f")
	km.Register(EventCursorHead, "home")
	km.Register(EventCursorHead, "c-a")
	km.Register(EventCursorEnd, "end")
	km.Register(EventCursorEnd, "c-e")
	km.Register(EventBackspaceCmdline, "c-h")
	km.Register(EventBackspaceCmdline, "backspace")
	km.Register(EventBackspaceCmdline, "backspace2")
	km.Register(EventDeleteCmdline, "delete")
	km.Register(EventDeleteWordCmdline, "c-w")
	km.Register(EventClearToHeadCmdline, "c-u")
	km.Register(EventClearCmdline, "c-k")
	km.Register(EventExitCmdline, "escape")
	km.Register(EventExitCmdline, "c-c")
	km.Register(EventExecuteCmdline, "enter")
	km.Register(EventExecuteCmdline, "c-j")
	km.Register(EventExecuteCmdline, "c-m")
	kms[ModeCmdline] = km
	return kms
}
