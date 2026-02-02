package storageManager

import "notifinetwork/localfusion/proto/types"

type Storage interface {
	GetStore() *KeyValueStore
}

type EphemeralStorage struct {
	Store  KeyValueStore
	Queues map[string]*Queue
}

func (s *EphemeralStorage) GetQueue(queueName string, storageType types.FusionStorageType) *Queue {
	queueId := deriveStorageId(storageType, queueName)
	queue, ok := s.Queues[queueId]

	if !ok {
		queue = NewQueue()
		s.Queues[queueId] = queue
	}

	return queue
}

func (s *EphemeralStorage) GetStore() *KeyValueStore {
	return &s.Store
}

func NewEphemeralStorage() *EphemeralStorage {
	return &EphemeralStorage{
		Store: KeyValueStore{
			data: make(map[string]VersionedValue),
		},
		Queues: make(map[string]*Queue),
	}
}

type PersistentStorage struct {
	Store KeyValueStore
}

func (s *PersistentStorage) GetStore() *KeyValueStore {
	return &s.Store
}

func NewPersistentStorage() *PersistentStorage {
	return &PersistentStorage{
		Store: KeyValueStore{
			data: make(map[string]VersionedValue),
		},
	}
}
