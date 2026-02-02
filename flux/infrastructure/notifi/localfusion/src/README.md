# Local Fusion

This service exposes several gRPC api's which essentially stub out the functionality of the real Fusion services. This will allow developers to locally run the fusion hosts with their parser sources.

## API's

The following API's are exposed by this service:

1. BlockchainManager
2. StorageManager
3. TODO: SubscriptionManager

## Configuration

The following environment variables can be used to configure the service

**Note:** Notifi uses an enumeration called `BlockchainType` to identify various blockchains. Each environment, such as production or staging, supports only one network for a specific blockchain type. For instance, in the production environment, the `BlockchainType ETHEREUM` refers to the Ethereum mainnet, while in the staging environment, it links to the Goerli test network. It's important to note that the codebase doesn't explicitly differentiate between testnet and mainnet; instead, it simply associates each blockchain type with the appropriate network for the current environment.

| Name                    | Description                                                                          | Default                              |
| ----------------------- | ------------------------------------------------------------------------------------ | ------------------------------------ |
| BLOCKCHAIN_MANAGER_PORT | The port to run the blockchain manager on                                            | "50051"                              |
| EPHEMERAL_STORAGE_PORT  | The port to run the storage manager on                                               | "50052"                              |
| PERSISTENT_STORAGE_PORT | The port to run the storage manager on                                               | "50053"                              |
| ETHEREUM_RPC_URL        | The RPC URL you wish to use for the BlockchainType ETHEREUM                          | "https://ethereum.publicnode.com"    |
| POLYGON_RPC_URL         | The RPC URL you wish to use for the BlockchainType POLYGON                           | "https://polygon-bor.publicnode.com" |
| OPTIMISM_RPC_URL        | The RPC URL you wish to use for the BlockchainType OPTIMISM                          | "https://optimism.publicnode.com"    |
| ARBITRUM_RPC_URL        | The RPC URL you wish to use for the BlockchainType ARBITRUM                          | "https://arbitrum.publicnode.com"    |
| BASE_RPC_URL            | The RPC URL you wish to use for the BlockchainType BASE                              | "https://base.publicnode.com"        |
| AVALANCHE_RPC_URL       | The RPC URL you wish to use for the BlockchainType AVALANCHE                         | "https://avalanche.publicnode.com"   |
| BNB_RPC_URL             | The RPC URL you wish to use for the BlockchainType BNB                               | "https://bsc.publicnode.com"         |
| CUSTOM_CHAIN_CONFIGS    | A JSON Array of `{"blockchainType": BlockchainType, "rpcUrl": "http://url.here.com"} | "[]"                                 |

## Running

This service can be run via the docker image which is located at `notifinetwork/localfusion`. If running via the docker image, make sure to expose the ports you wish to use, for example:

```bash
docker pull notifinetwork/localfusion:latest
docker run -p 50051:50051 -p 50052:50052 -p 50053:50053 -p 50054:50054 -p 50055:50055 -p 50056:50056 notifinetwork/localfusion:latest
```

Alternatively, you can run the service locally by cloning the repo and running the following command:

```bash
go run .
```

## Development

### Prerequisites

- Go: `brew install go`
- protobuf: `brew install protobuf`
- protoc-gen-go: `go get -u github.com/golang/protobuf/protoc-gen-go`
- protoc-gen-go-grpc: `go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc`

### Scripts

#### Generating the protobuf files

```bash
./generate-protos.sh

```

#### Building and pushing the docker image

Login to docker hub via the CLI or the docker desktop app, and then execute the following command.

```bash
./build-and-push.sh <semver>
```

e.g. `./build-and-push.sh 0.0.1`
