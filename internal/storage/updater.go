package storage

type Updater interface {
	UpdateStorage(...string) error
}
