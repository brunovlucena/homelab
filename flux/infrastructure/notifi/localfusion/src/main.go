package main

import (
	"fmt"
	"notifinetwork/localfusion/config"
	"notifinetwork/localfusion/proto/blockchain_manager"
	"notifinetwork/localfusion/proto/storage_manager"
	subscription_manager "notifinetwork/localfusion/proto/subscription_manager"
	grpcServices "notifinetwork/localfusion/services/grpc"

	"log"
	"net"

	"google.golang.org/grpc"
)

// func startBlockchainManagerGrpcService() {
// 	go func() {
// 		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Config.BLOCKCHAIN_MANAGER_PORT))
// 		if err != nil {
// 			log.Fatalf("blockchain manager service failed to listen: %v", err)
// 		}
// 		s := grpc.NewServer()
// 		blockchainManagerGrpcService := grpcServices.NewBlockchainManagerService()
// 		blockchain_manager.RegisterBlockchainManagerServer(s, blockchainManagerGrpcService)
// 		log.Printf("blockchain manager server listening at %v", lis.Addr())
// 		if err := s.Serve(lis); err != nil {
// 			log.Fatalf("failed to serve blockchain manager: %v", err)
// 		}
// 	}()
// }

func startEphemeralStorageGrpcService() {
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Config.EPHEMERAL_STORAGE_PORT))
		if err != nil {
			log.Fatalf("ephemeral storage service failed to listen: %v", err)
		}

		s := grpc.NewServer()
		ephemeralStorageGrpcService := grpcServices.NewEphemeralStorageService()
		storage_manager.RegisterFusionEphemeralStorageServer(s, ephemeralStorageGrpcService)
		log.Printf("ephemeral storage server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve ephemeral storage: %v", err)
		}

	}()
}

func startPersistentStorageGrpcService() {
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Config.PERSISTENT_STORAGE_PORT))
		if err != nil {
			log.Fatalf("persistent storage service failed to listen: %v", err)
		}

		s := grpc.NewServer()
		persistentStorageGrpcService := grpcServices.NewPersistentStorageService()
		storage_manager.RegisterFusionPersistentStorageServer(s, persistentStorageGrpcService)
		log.Printf("persistent storage server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve persistent storage: %v", err)
		}

	}()
}

func startFusionEvmGrpcService() {
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Config.EVM_RPC_PORT))
		if err != nil {
			log.Fatalf("fusion evm rpc service failed to listen: %v", err)
		}
		s := grpc.NewServer()
		fusionEvmRpcService := grpcServices.NewFusionEvmRpcService()
		blockchain_manager.RegisterFusionEvmRpcServer(s, fusionEvmRpcService)
		log.Printf("fusion evm rpc server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve fusion evm rpc: %v", err)
		}
	}()
}

func startFusionSolanaGrpcService() {
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Config.SOLANA_RPC_PORT))
		if err != nil {
			log.Fatalf("fusion solana rpc service failed to listen: %v", err)
		}
		s := grpc.NewServer()
		fusionSolanaRpcService := grpcServices.NewFusionSolanaRpcService()
		blockchain_manager.RegisterFusionSolanaRpcServer(s, fusionSolanaRpcService)
		log.Printf("fusion solana rpc server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve fusion solana rpc: %v", err)
		}
	}()
}

func startFusionSubscriptionsGrpcService() {
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Config.SUBSCRIPTION_MANAGER_PORT))
		if err != nil {
			log.Fatalf("fusion subscriptions service failed to listen: %v", err)
		}
		s := grpc.NewServer()
		fusionSubscriptionsService := grpcServices.NewFusionSubscriptionsService()
		subscription_manager.RegisterFusionSubscriptionsServer(s, fusionSubscriptionsService)
		log.Printf("fusion subscriptions server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve fusion subscriptions: %v", err)
		}
	}()
}

func main() {
	config.LoadFromEnv()
	// Start the blockchain manager service and don't close main thread
	//startBlockchainManagerGrpcService()
	startEphemeralStorageGrpcService()
	startPersistentStorageGrpcService()
	startFusionEvmGrpcService()
	startFusionSolanaGrpcService()
	startFusionSubscriptionsGrpcService()
	select {}
}
