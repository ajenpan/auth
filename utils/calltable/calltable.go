package calltable

import (
	"sync"
)

type CallTable struct {
	sync.RWMutex
	list map[string]*Method
}

func NewCallTable() *CallTable {
	return &CallTable{
		list: make(map[string]*Method),
	}
}

func (m *CallTable) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.list)
}

func (m *CallTable) Get(name string) *Method {
	m.RLock()
	defer m.RUnlock()

	ret, has := m.list[name]
	if has {
		return ret
	}
	return nil
}

func (m *CallTable) Range(f func(key string, value *Method) bool) {
	m.Lock()
	defer m.Unlock()
	for k, v := range m.list {
		if !f(k, v) {
			return
		}
	}
}

func (m *CallTable) Merge(other *CallTable, overWrite bool) int {
	ret := 0
	other.RWMutex.RLock()
	defer other.RWMutex.RUnlock()

	m.Lock()
	defer m.Unlock()

	for k, v := range other.list {
		_, has := m.list[k]
		if has && !overWrite {
			continue
		}
		m.list[k] = v
		ret++
	}
	return ret
}

func (m *CallTable) Add(name string, method *Method) bool {
	m.Lock()
	defer m.Unlock()
	if _, has := m.list[name]; has {
		return false
	}
	m.list[name] = method
	return true
}
