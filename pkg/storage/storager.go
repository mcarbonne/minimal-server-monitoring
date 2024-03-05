package storage

type Storager interface {
	Sync(force bool)
	Get(key string) (value string, exists bool)
	Set(key, value string) (changed bool)
	Remove(key string)
}
