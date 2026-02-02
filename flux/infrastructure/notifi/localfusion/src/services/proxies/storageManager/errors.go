package storageManager

import "fmt"

type VersionAlreadyExistsError struct {
	Key     string
	Version uint64
}

func (e *VersionAlreadyExistsError) Error() string {
	return "Version already exists for key " + e.Key + " with version " + fmt.Sprintf("%d", e.Version)
}
