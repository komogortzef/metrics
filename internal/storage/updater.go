package storage

type StorageUpdater interface {
	UpdateStorage(string, string, string) error
}
