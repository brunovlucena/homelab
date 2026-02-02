package storageManager

import (
	"notifinetwork/localfusion/proto/types"
)

var defaultVersion uint64 = 0

type VersionedValue struct {
	Version uint64
	Value   string
}

type KeyValueStore struct {
	data map[string]VersionedValue
}

func (d *KeyValueStore) Get(storageType types.FusionStorageType, key string) *VersionedValue {
	keyString := deriveStorageId(storageType, key)
	value, ok := d.data[keyString]

	if !ok {
		return nil
	}

	return &value
}

func (d *KeyValueStore) Put(contextId string, storageType types.FusionStorageType, key string, value string, version *uint64) error {
	keyString := deriveStorageId(storageType, key)
	existingValue, ok := d.data[keyString]

	if version == nil {
		version = &defaultVersion
	}

	// If the key doesn't exist or the version is different, update the value
	if !ok || existingValue.Version != *version {
		d.data[keyString] = VersionedValue{
			Version: *version,
			Value:   value,
		}

		return nil
	}

	return &VersionAlreadyExistsError{
		Key:     keyString,
		Version: existingValue.Version,
	}
}

func (d *KeyValueStore) Delete(contextId string, storageType types.FusionStorageType, key string) {
	keyString := deriveStorageId(storageType, key)
	delete(d.data, keyString)
}
