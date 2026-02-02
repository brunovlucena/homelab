rm -rf proto
mkdir proto
protoc --go_out=./proto --go-grpc_out=./proto --proto_path=../../../Protos \
  ../../../Protos/notifi/common/v1/types.proto \
  ../../../Protos/services/blockchain_manager/v1/blockchain_manager.proto \
  ../../../Protos/services/storage_manager/v1/storage_manager.proto \
  ../../../Protos/services/subscription_manager/v1/subscription_manager.proto \
  ../../../Protos/services/scheduler/v1/scheduler.proto \
  ../../../Protos/services/user_manager/v1/user_manager.proto
cp -rf ./proto/notifinetwork/localfusion/proto ./
rm -rf ./proto/notifinetwork
