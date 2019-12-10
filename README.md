## Platform Take-home Challenge

**Tasks**

1. create a local running Kubernetes cluster with the tooling of your choice [DONE]
2. Install Postgres on the cluster. [DONE]
3. Create and deploy a basic application in Go and connect to Postgres [DONE]
4. Perform simple CRUD operations [DONE]
5. Load-test your application. [DONE] 
6. Setup basic monitoring and logging for the application with tools of your choice. [DONE]

### TD;DR

**How to Test**

```sh
go get github.com/brunovlucena/homelab
./start.sh
# Edit crud.sh
make crud
make load-test
```

or

```sh
make pre-install
make bootstrap-cluster
make helm-install
make tunnel-registry
make build-deploy-operator 
make deploy-myapp-test 
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
load-test                      Run Load Tests
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

**NOTE**: You should edit `/etc/hosts` ([minikube_ip] kibana.local grafana.local prom.local myapp.local)

- [Kibana](http://kibana.local/kibana)
- [Grafana](http://grafana.local)
- [Prometheus-Monitoring](http://prom.local:30100)
- [Prometheus-Rook](http://prom.local:30200)
- [Prometheus-Dev](http://prom.local:30300)



## API In Golang

API in golang using Chi Router 

- Tested: minikube v1.5.1 and Kubernetes: v1.16.2


### Non-Functional Requirements

1. Centralized configuration (TODO)
2. Service Discovery (TODO)
3. Logging
4. Distributed Tracing (TODO)
5. Circuit Breaking (TODO)
6. Load balancing (TODO)
7. Monitoring
8. Security (TODO)


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

- The **application** servers the API on the port defined by the environment variable `API_CONTAINER_PORT`.

- **Database Variables**:

| Variable | Type | Example | Description |
| -------- | ---- | ------- | ----------- |
|`DATABASE_TYPE`| string | "postgres"			| ## Database type
|`DATABASE_HOST`| string | "postgres.storage"	| ## Database location
|`DATABASE_PORT`| string | "5432"				| ## Port
|`DATABASE_USER`| string | "postgres"			| ## User
|`DATABASE_PASS`| string | "postgres"			| ## Pass
|`DATABASE_NAME`| string | "myapp"				| ## Database name

### Operator (apps/app-example/deploy/chart/myapp/crds/myapp.yaml)

```yaml
apiVersion: myapp.com/v1alpha1
kind: MyApp
metadata:
  name: myapp
spec:
  size: 3
  database_type: postgres
  database_name: myapp
  host: postgres.storage
  port: 5432
  user: postgres
  pass: postgres
  container_port: 8000
  watch:
    - namespace: storage
      deployments:
        - name: postgres
```


## TODO

- Search
- Ingress TLS
- Make Operator interact with cluster (E.g create a backup given some alert).
- Grafana Loki
- Fix Velero's Job
- Use Vault for secrets

