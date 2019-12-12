## Home Lab

**Tasks**

1. create a local running Kubernetes cluster with the tooling of your choice [DONE]
2. Install Postgres on the cluster. [DONE]
3. Create and deploy a basic application in Go and connect to Postgres [DONE]
4. Perform simple CRUD operations [DONE]
5. Load-test your application. [DONE] 
6. Setup basic monitoring and logging for the application with tools of your choice. [DONE]

**TODO**

- Search
- Ingress TLS
- Make Operator interact with cluster (E.g create a backup given some alert).
- Grafana Loki
- Fix Velero's Job
- Use Vault for secrets

### TD;DR

**How to Test**

```sh
go get github.com/brunovlucena/homelab
./start.sh
# Edit crud.sh to change tests
make crud
make load-test
make test ## Find and FindAll failing because the http.StatusCode but they work!
make test-gui
```

**Make Commands**

```sh
bootstrap-cluster              Bootstraps cluster (E.g. make bootstrap).
bootstrap-operator             Builds operator (E.g. make bootstrap-operator).
build-deploy-operator          Deploys operator (E.g. make build-deploy-operator).
build-myapp                    Builds binary app example (E.g. make build-myapp).
check-pod-security             outputs infomation about the cluster
clean-cluster                  Cleans Minikube (E.g. make clean-cluster).
debug-myapp                    Runs dlv (E.g. make debug-myapp).
debug-myapp-tests              Runs dlv to debug Tests.
deploy-myapp-test              Builds image for app example (E.g. make build-push-myapp latest).
helm-install                   Installs components via helm charts.
help                           Help.
load-test                      Run Load Tests (E.g make load-test)
pre-install                    Pre-Installs tools (E.g: $ make pre-install).
run-myapp                      Runs app example on host (E.g. make run-myapp).
run-postgres-local             Runs postgres on host (E.g. make run-postgres-local).
skaffold                       Uses skaffold during the development
sniff                          Sniffs comunication (E.g. make sniff)
start-cluster                  Starts cluster.
stop-cluster                   Stops cluster.
test-gui                       Run Go Tests (Browser)
test                           Run Go Tests
tunnel-registry                Creates a tunnel to minikube's registry (E.g. make tunnel-registry).
```


### Infra Endpoints

**NOTE**: You should edit `/etc/hosts` ([minikube_ip] kibana.local grafana.local dashboard.local prom.local myapp.local)

- [Kibana](http://kibana.local/kibana)
- [Grafana](http://grafana.local)
- [Dashboard](http://dashboard.local)
- [Prometheus-Monitoring](http://prom.local:30100)
- [Prometheus-Rook](http://prom.local:30200)
- [Prometheus-Dev](http://prom.local:30300)
- [RabbitMQ](http://rabbitmq.local/)


### API In Golang

API in golang using Chi Router 

- Tested: minikube v1.5.1 and Kubernetes: v1.16.2


#### Non-Functional Requirements

1. Centralized configuration (TODO)
2. Service Discovery (TODO)
3. Logging
4. Distributed Tracing (TODO)
5. Circuit Breaking (TODO)
6. Load balancing (TODO)
7. Monitoring
8. Security (TODO)


#### Endpoints

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



#### Configuration

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

### Postgres HA Cluster

**Questions**

- How do I automatically deploy a new PostgreSQL instance?
- How do I failover a PostgreSQL pod to another availability zone if my PostgreSQL instance goes down?
- How do I resize my PostgreSQL volume if I am running out of space?
- How do I snapshot and backup PostgreSQL for disaster recovery?
- How do I test upgrades?
- Can I take my PostgreSQL deployment and run it in any environment if needed?

**Achieving HA with PostgreSQL**

- [Zalando Operator](https://github.com/zalando/postgres-operator)
