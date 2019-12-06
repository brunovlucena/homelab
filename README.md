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
make operator-build
make operator-deploy
make image-build-myapp
```

**Make Commands**

```sh
bootstrap-cluster              Bootstraps cluster (E.g. make bootstrap).
build-myapp                    Builds app example (E.g. make build-myapp).
clean-cluster                  Cleans Minikube (E.g. make clean-cluster).
helm-install                   Installs components via helm charts.
help                           Help. 
image-build-myapp              Builds image for app example (E.g. make image-build-myapp).
mod-tidy-myapp                 Runs app example (E.g. make mod-tidy-myapp).
operator-build                 Builds operator (E.g. make operator-build).
operator-deploy                Deploys operator (E.g. make operator-deploy).
pre-install                    Pre-Installs tools (E.g: $ make pre-install).
run-myapp                      Runs app example (E.g. make run-myapp).
start-cluster                  Starts cluster.
stop-cluster                   Stops cluster.
test-operator                  Tests operator (E.g. make test-operator).
tunnel-registry                Creates a tunnel to minikube's registry (E.g. make tunnel-registry).
```

