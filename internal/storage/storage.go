package storage

type Storage interface {
	Save(data ...[]byte) error
	Fetch(keys ...string) (any, error)
}
