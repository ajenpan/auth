package calltable

import (
	"reflect"
	"sync"
)

type MethodStyle int

const (
	StyleMicro   = iota // func (context.Context, proto.Message, proto.Message) ( error)
	StyleGRpc    = iota // func (context.Context, proto.Message) (proto.Message, error)
	StyleAsync   = iota // func (session, proto.Message) error
	StyleRequest = iota // func (session, proto.Message) (proto.Message, error)
)

type NewTypeFunc = func() interface{}

const EndpointSplit = "/"

type Method struct {
	Imp   reflect.Method
	Style MethodStyle

	H interface{}

	RequestType  reflect.Type
	ResponseType reflect.Type

	reqPool  *sync.Pool
	respPool *sync.Pool
}

func (m *Method) InitPool() {
	if m.RequestType != nil {
		m.reqPool = &sync.Pool{New: m.NewRequest}
	}

	if m.ResponseType != nil {
		m.respPool = &sync.Pool{New: m.NewResponse}
	}
}

func (m *Method) Call(args ...interface{}) []reflect.Value {
	values := make([]reflect.Value, 0, len(args)+1)
	values = append(values, reflect.ValueOf(m.H))
	for _, v := range args {
		values = append(values, reflect.ValueOf(v))
	}
	return m.Imp.Func.Call(values)
}

func (m *Method) NewRequest() interface{} {
	if m.RequestType == nil {
		return nil
	}
	return reflect.New(m.RequestType).Interface()
}

func (m *Method) NewResponse() interface{} {
	if m.ResponseType == nil {
		return nil
	}
	return reflect.New(m.ResponseType).Interface()
}

func (m *Method) GetRequest() interface{} {
	if m.reqPool == nil {
		return m.NewRequest()
	}
	return m.reqPool.Get()
}

func (m *Method) PutRequest(req interface{}) {
	if m.reqPool == nil {
		return
	}
	m.reqPool.Put(req)
}

func (m *Method) GetResponse() interface{} {
	if m.respPool == nil {
		return m.NewResponse()
	}
	return m.respPool.Get()
}

func (m *Method) PutResponse(resp interface{}) {
	if m.respPool == nil {
		return
	}
	m.respPool.Put(resp)
}

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
