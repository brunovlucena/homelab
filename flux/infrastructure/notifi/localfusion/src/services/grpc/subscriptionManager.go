package grpcServices

import (
	context "context"
	subscription_manager "notifinetwork/localfusion/proto/subscription_manager"
)

type FusionSubscriptionsService struct {
	subscription_manager.UnimplementedFusionSubscriptionsServer
}

func NewFusionSubscriptionsService() *FusionSubscriptionsService {
	return &FusionSubscriptionsService{}
}

func (s *FusionSubscriptionsService) GetSubscriptions(ctx context.Context, req *subscription_manager.GetSubscriptionsRequest) (*subscription_manager.GetSubscriptionsResponse, error) {
	// TODO: Implement actual logic. Maybe return a static list from a local csv?
	return &subscription_manager.GetSubscriptionsResponse{}, nil
}
