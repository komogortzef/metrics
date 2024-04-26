package storage

type Updater interface {
	UpdateStorage(string, string, string) error
}
