package domain

type StorageProvider interface {
	Upload(filename string, data []byte) (string, error)
}
