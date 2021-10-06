package mngconn

import (
	"strconv"
	"sync"
)

type handlerMap struct {
	m sync.Map
}

func (h *handlerMap) Add(seq int, value handlerType) {
	h.m.Store(strconv.Itoa(seq), value)
}

func (h *handlerMap) GetAndRemove(seq int) (handlerType, bool) {
	v, ok := h.m.LoadAndDelete(strconv.Itoa(seq))

	if !ok {
		return nil, false
	}

	return v.(handlerType), true
}

func (h *handlerMap) Remove(seq int) {
	h.m.Delete(strconv.Itoa(seq))
}
