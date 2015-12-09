package ttlslicemap

import (
	"sync"
	"time"
)

type RWLocker interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type itemSlice struct {
	key    string
	parent *TTLSliceMap
	timer  *time.Timer
	items  []interface{}
	mutex  sync.Locker
}

func newItemSlice(key string, parent *TTLSliceMap) *itemSlice {
	is := &itemSlice{
		key:    key,
		parent: parent,
		mutex:  new(sync.Mutex),
	}

	is.timer = time.AfterFunc(
		is.parent.ttl,
		func() {
			is.parent.removeIdentical(is.key, is)
		},
	)

	return is
}

func (is *itemSlice) refresh() {
	is.timer.Reset(is.parent.ttl)
}

func (is *itemSlice) add(item interface{}) {
	is.mutex.Lock()
	defer is.mutex.Unlock()

	is.items = append(is.items, item)
	is.refresh()
}

func (is *itemSlice) get() (items []interface{}) {
	is.mutex.Lock()
	defer is.mutex.Unlock()

	is.refresh()

	items = is.items
	return
}

type TTLSliceMap struct {
	ttl          time.Duration
	itemSliceMap map[string]*itemSlice
	mutex        RWLocker
}

func New(ttl time.Duration) *TTLSliceMap {
	tsm := &TTLSliceMap{
		ttl:          ttl,
		itemSliceMap: make(map[string]*itemSlice),
		mutex:        new(sync.RWMutex),
	}

	return tsm
}

func (tsm *TTLSliceMap) Add(key string, item interface{}) (new bool) {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()

	itemSlice, exists := tsm.itemSliceMap[key]
	if exists {
		new = false
	} else {
		new = true
		itemSlice = newItemSlice(key, tsm)
	}

	itemSlice.add(item)
	tsm.itemSliceMap[key] = itemSlice

	return
}

func (tsm *TTLSliceMap) Get(key string) (items []interface{}, exists bool) {
	tsm.mutex.RLock()
	defer tsm.mutex.RUnlock()

	itemSlice, exists := tsm.itemSliceMap[key]
	if !exists {
		return nil, false
	}

	items = itemSlice.get()
	return
}

func (tsm *TTLSliceMap) Remove(key string) (existed bool) {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()

	_, existed = tsm.itemSliceMap[key]
	delete(tsm.itemSliceMap, key)

	return
}

func (tsm *TTLSliceMap) removeIdentical(key string, itemSlice *itemSlice) {
	tsm.mutex.Lock()
	defer tsm.mutex.Unlock()

	candidate, exists := tsm.itemSliceMap[key]
	if exists && candidate == itemSlice {
		delete(tsm.itemSliceMap, key)
	}
}

func (tsm *TTLSliceMap) Count() int {
	tsm.mutex.RLock()
	defer tsm.mutex.RUnlock()

	return len(tsm.itemSliceMap)
}
