package grpcServices

import (
	"context"
	proto "notifinetwork/localfusion/proto/storage_manager"
	storageManager "notifinetwork/localfusion/services/proxies/storageManager"
)

func Put(storage storageManager.Storage, context context.Context, request *proto.FusionPutStringRequest) (*proto.FusionPutStringResponse, error) {
	error := storage.GetStore().Put(request.ContextId, request.StorageType, request.Key, request.Value, request.Version)

	if error != nil {
		return nil, error
	}

	return &proto.FusionPutStringResponse{}, nil
}

func Get(storage storageManager.Storage, ctx context.Context, req *proto.FusionStorageGetRequest) (*proto.FusionStorageGetResponse, error) {
	response := &proto.FusionStorageGetResponse{
		Values: make(map[string]*proto.FusionStorageGetResponse_Record),
	}

	for _, key := range req.Keys {
		value := storage.GetStore().Get(req.StorageType, key)
		if value != nil {
			response.Values[key] = &proto.FusionStorageGetResponse_Record{
				Value:   value.Value,
				Version: value.Version,
			}
		}
	}

	return response, nil
}

func Delete(storage storageManager.Storage, ctx context.Context, req *proto.FusionStorageDeleteRequest) (*proto.FusionStorageDeleteResponse, error) {
	storage.GetStore().Delete(req.ContextId, req.StorageType, req.Key)
	return &proto.FusionStorageDeleteResponse{}, nil
}

// Ephemeral Storage Manager

type EphemeralStorageService struct {
	proto.UnimplementedFusionEphemeralStorageServer
	storageManager *storageManager.EphemeralStorage
}

func (s *EphemeralStorageService) Put(ctx context.Context, req *proto.FusionPutStringRequest) (*proto.FusionPutStringResponse, error) {
	return Put(s.storageManager, ctx, req)
}

func (s *EphemeralStorageService) Get(ctx context.Context, req *proto.FusionStorageGetRequest) (*proto.FusionStorageGetResponse, error) {
	return Get(s.storageManager, ctx, req)
}

func (s *EphemeralStorageService) Delete(ctx context.Context, req *proto.FusionStorageDeleteRequest) (*proto.FusionStorageDeleteResponse, error) {
	return Delete(s.storageManager, ctx, req)
}

func (s *EphemeralStorageService) Peek(ctx context.Context, req *proto.FusionStoragePeekRequest) (*proto.FusionStoragePeekResponse, error) {
	queue := s.storageManager.GetQueue(req.QueueName, req.StorageType)
	value := queue.Peek()

	if value != nil {
		return &proto.FusionStoragePeekResponse{
			Value: *value,
		}, nil
	}

	return &proto.FusionStoragePeekResponse{}, nil
}

func (s *EphemeralStorageService) Enqueue(ctx context.Context, req *proto.FusionStorageEnqueueRequest) (*proto.FusionStorageEnqueueResponse, error) {
	queue := s.storageManager.GetQueue(req.QueueName, req.StorageType)
	queue.Enqueue(req.Value)

	return &proto.FusionStorageEnqueueResponse{}, nil
}

func (s *EphemeralStorageService) Dequeue(ctx context.Context, req *proto.FusionStorageDequeueRequest) (*proto.FusionStorageDequeueResponse, error) {
	queue := s.storageManager.GetQueue(req.QueueName, req.StorageType)
	value := queue.Dequeue()
	if value != nil {
		return &proto.FusionStorageDequeueResponse{
			Value: *value,
		}, nil
	}

	return &proto.FusionStorageDequeueResponse{}, nil
}

func NewEphemeralStorageService() *EphemeralStorageService {
	storageManager := storageManager.NewEphemeralStorage()
	return &EphemeralStorageService{
		storageManager: storageManager,
	}
}

// Persistent Storage Manager

type PersistentStorageService struct {
	proto.UnimplementedFusionPersistentStorageServer
	storageManager *storageManager.PersistentStorage
}

func (s *PersistentStorageService) Put(ctx context.Context, req *proto.FusionPutStringRequest) (*proto.FusionPutStringResponse, error) {
	return Put(s.storageManager, ctx, req)
}

func (s *PersistentStorageService) Get(ctx context.Context, req *proto.FusionStorageGetRequest) (*proto.FusionStorageGetResponse, error) {
	return Get(s.storageManager, ctx, req)
}

func (s *PersistentStorageService) Delete(ctx context.Context, req *proto.FusionStorageDeleteRequest) (*proto.FusionStorageDeleteResponse, error) {
	return Delete(s.storageManager, ctx, req)
}

func NewPersistentStorageService() *PersistentStorageService {
	storageManager := storageManager.NewPersistentStorage()
	return &PersistentStorageService{
		storageManager: storageManager,
	}
}
