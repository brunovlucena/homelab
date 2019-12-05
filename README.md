## Platform Take-home Challenge

**Tasks**

1. create a local running Kubernetes cluster with the tooling of your choice
2. Install Postgres on the cluster.
3. Create and deploy a basic application in Go and connect to Postgres
4. Perform simple CRUD operations
5. Load-test your application.
6. Setup basic monitoring and logging for the application with tools of your choice.

### TD;DR

**How to Test**

```sh
make pre-install
make bootstrap
```

**Make Commands**

```sh
bootstrap-cluster              Bootstraps cluster (E.g. make bootstrap).
clean-cluster                  Cleans Minikube (E.g. make clean-cluster).
help                           Help. 
pre-install                    Pre-Installs tools (E.g: $ make pre-install).
```
