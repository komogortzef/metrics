package storage

type Storage interface {
	Save(data ...[]byte) error
}
