## Platform Take-home Challenge

**Tasks**

1. create a local running Kubernetes cluster with the tooling of your choice [DONE]
2. Install Postgres on the cluster.
3. Create and deploy a basic application in Go and connect to Postgres
4. Perform simple CRUD operations
5. Load-test your application.
6. Setup basic monitoring and logging for the application with tools of your choice.

### TD;DR

**How to Test**

```sh
./start.sh
```

or

```sh
make pre-install
make bootstrap-cluster
make helm-install
make tunnel-registry
build-deploy-operator-test 
```

**Make Commands**

```sh
bootstrap-cluster              Bootstraps cluster (E.g. make bootstrap).
bootstrap-operator             Builds operator (E.g. make bootstrap-operator).
build-deploy-myapp             Builds image for app example (E.g. make build-push-myapp latest).
build-deploy-operator          Deploys operator (E.g. make build-deploy-operator).
build-deploy-operator-test     Tests MyAppOperator (E.g. make build-deploy-test). 
build-myapp                    Builds binary app example (E.g. make build-myapp).
check-pod-security             outputs infomation about the cluster
clean-cluster                  Cleans Minikube (E.g. make clean-cluster).
debug-myapp                    Runs dlv (E.g. make debug-myapp).
helm-install                   Installs components via helm charts.
help                           Help. 
load-test                      Run Go Tests
pre-install                    Pre-Installs tools (E.g: $ make pre-install).
run-myapp                      Runs app example on host (E.g. make run-myapp).
run-postgres-local             Runs postgres on host (E.g. make run-postgres-local).
skaffold                       Uses skaffold during the development
sniff                          Sniffs comunication (E.g. make sniff)
start-cluster                  Starts cluster.
stop-cluster                   Stops cluster.
test                           Run Go Tests
tunnel-registry                Creates a tunnel to minikube's registry (E.g. make tunnel-registry).
```

## Infra Endpoints

**NOTE**: You should edit `/etc/hosts` to point minikube's ip

[Kibana](http://kibana.local/kibana)
[Grafana](http://grafana.local)
[Prometheus-Monitoring](http://prom.local:30100)
[Prometheus-Rook](http://prom.local:30200)
[Prometheus-Dev](http://prom.local:30300)


## API In Golang

API in golang using Chi Router 

- Tested: minikube v1.5.1 and Kubernetes: v1.16.2


### Non-Functional Requirements

1. Centralized configuration
2. Service Discovery
3. Logging
4. Distributed Tracing
5. Circuit Breaking
6. Load balancing
7. Edge
8. Monitoring
7. Security



### Endpoints

| Name   | Method      | URL
| ---    | ---         | ---
| List   | `GET`       | `/configs`
| Create | `POST`      | `/configs`
| Get    | `GET`       | `/configs/{name}`
| Update | `PUT/PATCH` | `/configs/{name}`
| Delete | `DELETE`    | `/configs/{name}`
| Query  | `GET`       | `/search?metadata.key=value`


#### Query

The query endpoint **MUST** return all configs that satisfy the query argument.

Query example-1:

```sh
curl http://localhost:8000/configs/search?metadata.monitoring.enabled=true
```

Response example:

```json
[
  {
    "name": "foo",
    "metadata": {
      "monitoring": {
        "enabled": "true"
      },
      "limits": {
        "cpu": {
          "enabled": "false",
          "value": "300m"
        }
      }
    }
  },
  {
    "name": "bar",
    "metadata": {
      "monitoring": {
        "enabled": "true"
      },
      "limits": {
        "cpu": {
          "enabled": "true",
          "value": "250m"
        }
      }
    }
  },
]
```


#### Schema

- **Config**
  - Name (string)
  - Metadata (nested key:value pairs where both key and value are strings of arbitrary length)


### Configuration

The application servers the API on the port defined by the environment variable `API_CONTAINER_PORT`.

