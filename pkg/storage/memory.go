package storage

import "sync"

type MemoryStorage struct {
	database        map[string]string
	unsyncedChanges bool
	mutex           sync.Mutex
}

func NewMemoryStorage() Storager {
	return &MemoryStorage{
		database:        make(map[string]string),
		unsyncedChanges: false,
	}
}

func (ms *MemoryStorage) Sync(force bool) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.syncUnsafe(force)
}

func (ms *MemoryStorage) syncUnsafe( /*force*/ bool) {
	ms.unsyncedChanges = false
}

func (ms *MemoryStorage) Get(key string) (string, bool) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	value, exists := ms.database[key]
	return value, exists
}

func (ms *MemoryStorage) Set(key, value string) bool {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	if ms.database[key] != value {
		ms.unsyncedChanges = true
		ms.database[key] = value
		return true
	} else {
		return false
	}
}

func (ms *MemoryStorage) Remove(key string) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	if _, ok := ms.database[key]; ok {
		delete(ms.database, key)
		ms.unsyncedChanges = true
	}
}
