package storage

type Storage interface {
	Save(...string) error
	Retrive() (string, error)
}
