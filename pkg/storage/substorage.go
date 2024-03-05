package storage

type SubStorage struct {
	underlyingStorage Storager
	prefix            string
}

func NewSubStorage(underlyingStorage Storager, prefix string) Storager {
	return &SubStorage{
		underlyingStorage: underlyingStorage,
		prefix:            prefix,
	}
}

func (substorage *SubStorage) Sync(force bool) {
	substorage.underlyingStorage.Sync(force)
}

func (substorage *SubStorage) Get(key string) (value string, exists bool) {
	return substorage.underlyingStorage.Get(substorage.prefix + key)
}

func (substorage *SubStorage) Set(key, value string) (changed bool) {
	return substorage.underlyingStorage.Set(substorage.prefix+key, value)
}

func (substorage *SubStorage) Remove(key string) {
	substorage.underlyingStorage.Remove(substorage.prefix + key)
}
