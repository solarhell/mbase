package mycontext

import (
	"net/netip"
	"reflect"
	"sync"
	"time"
)

// Env should be used as a map, but it has an inherited mode, overlay liked.
// If you set, the value should be stored at local.
// If you get, the value should be got from local first, otherwise from parent
type Env interface {
	// Fork returns an inherited sub Env, and I am its parent
	Fork() Env

	// Set always set key & value at local storage
	Set(key, value interface{})
	// Get always check local, if the key not exists, then check parent
	Get(key interface{}) (value interface{}, ok bool)
	Has(key interface{}) (ok bool)
	Keys() []interface{}

	GetInt(key interface{}) (int, bool)
	GetInt64(key interface{}) (int64, bool)
	GetUint(key interface{}) (uint, bool)
	GetUint64(key interface{}) (uint64, bool)
	GetFloat64(key interface{}) (float64, bool)
	GetBool(key interface{}) (bool, bool)
	GetString(key interface{}) (string, bool)
	GetNetIPAddr(key interface{}) (netip.Addr, bool)
	GetNetIPAddrPort(key interface{}) (netip.AddrPort, bool)
	GetNetIPPrefix(key interface{}) (netip.Prefix, bool)
	GetTime(key interface{}) (time.Time, bool)
	GetDuration(key interface{}) (time.Duration, bool)
}

type env struct {
	parent *env

	vals sync.Map
}

var _ Env = &env{}

// NewEnv return a simple Env, use sync.Map as it's storage
func NewEnv() Env {
	return &env{}
}

func (e *env) fork() *env {
	return &env{
		parent: e,
	}
}

func (e *env) Fork() Env {
	return e.fork()
}

func (e *env) Set(key, value interface{}) {
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}
	e.vals.Store(key, value)
}

func (e *env) Get(key interface{}) (value interface{}, ok bool) {
	// from local
	if value, ok := e.vals.Load(key); ok {
		return value, ok
	}
	// otherwise from parent
	if e.parent != nil {
		return e.parent.Get(key)
	}
	return nil, false
}

func (e *env) Has(key interface{}) (ok bool) {
	_, ok = e.Get(key)
	return
}

func (e *env) keys() map[interface{}]struct{} {
	var keys map[interface{}]struct{}
	if e.parent != nil {
		keys = e.parent.keys()
	} else {
		keys = make(map[interface{}]struct{})
	}
	e.vals.Range(func(k, v interface{}) bool {
		keys[k] = struct{}{}
		return true
	})
	return keys
}

func (e *env) Keys() []interface{} {
	var (
		idx  = e.keys()
		keys = make([]interface{}, 0, len(idx))
	)
	for k := range idx {
		keys = append(keys, k)
	}
	return keys
}

func (e *env) GetInt(key interface{}) (int, bool) {
	if value, ok := e.Get(key); ok {
		return value.(int), true
	}
	return 0, false
}

func (e *env) GetInt64(key interface{}) (int64, bool) {
	if value, ok := e.Get(key); ok {
		return value.(int64), true
	}
	return 0, false
}

func (e *env) GetUint(key interface{}) (uint, bool) {
	if value, ok := e.Get(key); ok {
		return value.(uint), true
	}
	return 0, false
}

func (e *env) GetUint64(key interface{}) (uint64, bool) {
	if value, ok := e.Get(key); ok {
		return value.(uint64), true
	}
	return 0, false
}

func (e *env) GetFloat64(key interface{}) (float64, bool) {
	if value, ok := e.Get(key); ok {
		return value.(float64), true
	}
	return 0, false
}

func (e *env) GetBool(key interface{}) (bool, bool) {
	if value, ok := e.Get(key); ok {
		return value.(bool), true
	}
	return false, false
}

func (e *env) GetString(key interface{}) (string, bool) {
	if value, ok := e.Get(key); ok {
		return value.(string), true
	}
	return "", false
}

func (e *env) GetNetIPAddr(key interface{}) (netip.Addr, bool) {
	if value, ok := e.Get(key); ok {
		return value.(netip.Addr), true
	}
	return netip.Addr{}, false
}

func (e *env) GetNetIPAddrPort(key interface{}) (netip.AddrPort, bool) {
	if value, ok := e.Get(key); ok {
		return value.(netip.AddrPort), true
	}
	return netip.AddrPort{}, false
}

func (e *env) GetNetIPPrefix(key interface{}) (netip.Prefix, bool) {
	if value, ok := e.Get(key); ok {
		return value.(netip.Prefix), true
	}
	return netip.Prefix{}, false
}

func (e *env) GetTime(key interface{}) (time.Time, bool) {
	if value, ok := e.Get(key); ok {
		return value.(time.Time), true
	}
	return time.Time{}, false
}

func (e *env) GetDuration(key interface{}) (time.Duration, bool) {
	if value, ok := e.Get(key); ok {
		return value.(time.Duration), true
	}
	return 0, false
}
