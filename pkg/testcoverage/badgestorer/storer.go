package badgestorer

type Storer interface {
	Store(data []byte) (hasUpdated bool, err error)
}
