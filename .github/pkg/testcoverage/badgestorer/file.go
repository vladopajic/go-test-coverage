package badgestorer

import "os"

type fileStorer struct {
	filename string
}

func NewFile(filename string) Storer {
	return &fileStorer{filename: filename}
}

//nolint:gosec,mnd,wrapcheck // relax
func (s *fileStorer) Store(data []byte) (bool, error) {
	err := os.WriteFile(s.filename, data, 0o644)
	if err != nil {
		return false, err
	}

	return true, nil
}
