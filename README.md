## Home Lab

Create a local development environment for Kubernetes.

- (Virtualbox) Tested: minikube v1.5.1 and Kubernetes: v1.17.0
- (Kind) Tested: minikube v1.5.1 and Kubernetes: v1.16.3

**TODO**:

- Add Grafana Loki
- Fix Velero/Minio
- Fix Rook (kind and Minikube)
- Fix gogs
- Fix gocd
- Fix Redis' Slave
- Add Postgres HA
- Add MySQL HA

**Basic Configuration(in Makefile)**:

| Variable | Value |
| -------- | ------- |
|`MINIKUBE_VERSION` |   v1.5.2
|`CLUSTER_CPUS`     |   6
|`CLUSTER_MEMORY`   |   4096mb 
|`CLUSTER_DISK`     |   20GB
|`CLUSTER_DISK_EXTRA` | 15GB
|`CLUSTER_VERSION`  |   v1.17.0
|`VM_DRIVER`        |   none or virtualbox
|`KIND_VERSION`	    |   v0.6.1
|`CLUSTER_NAME`     |   homelab

**Components(in Makefile)**:

| Variable | Value |
| -------- | ------- |
|`CNI`      |	calico
|`MESH`     |	linkerd
|`BASIC`    |	enabled
|`MONITORING` |	enabled
|`STORAGE`  |	disabled
|`CICD`     |	disabled
|`SECURITY` |	disabled
|`TESTING`  |	disabled
|`ROOK_CEPH`|	disabled
|`BACKUP`   |	disabled

**Tools(in Makefile)**:

| Variable | Value |
| -------- | ------- |
|`K9S_VERSION`      |	0.10.8
|`KUBECTL_VERSION`  |	v1.17.0
|`HELM_VERSION`     |	v3.0.1
|`SQUASH_VERSION`   |	v0.5.18
|`SONOBUOY_VERSION` |	0.16.1
|`GO_VERSION`       |	1.13.5
|`LINKERD_VERSION`  |	2.6.1
|`KREW_VERSION`     |   v0.3.3

**How to Test**:

```sh
go get github.com/brunovlucena/homelab
./start.sh
```

**Examples to Test**: 

- [guestbook-go-operator](https://github.com/brunovlucena/)
- [guestbook-go](https://github.com/brunovlucena/)
- [rest-api-go-async](https://github.com/brunovlucena/)
- [rest-api-go](https://github.com/brunovlucena/)


**Make Commands**:

```sh
add                            Adds components to the cluster.
bootstrap-cluster              Bootstraps a kubernetes cluster.
check-pod-security             outputs infomation about the cluster
help                           helper. 
pre-install                    pre-installs all nescessary tools to bootstrap and manage cluster.
sniff                          Sniffs comunication (E.g. make sniff)
start-cluster-kind             Starts cluster using Kind
start-cluster-minikube         Starts using minikube cluster.
stop-cluster                   Stops cluster.
tunnel                         Creates a tunnel to minikube's registry.
```

## BUGS

- [Rook OSD fails on Restart](https://github.com/rook/rook/issues/3289)


## Infra Endpoints

**Linux**: Update the file /etc/resolvconf/resolv.conf.d/base to have the following contents
```
search local
nameserver 192.168.99.100 # minikube'ip
timeout 5
```

**Mac OS**: Create a file in /etc/resolver/minikube-profilename-test
```
domain test
nameserver 192.168.99.100
search_order 1
timeout 5
```

- [Dashboard](http://dashboard.local)
- [Grafana](http://grafana.local)
- [Kibana](http://kibana.local/kibana)
- [RabbitMQ](http://rabbitmq.local/)
- [Prometheus-Kube-System](http://prom.local:30002)
- [Prometheus-Monitoring](http://prom.local:30001)
- [Prometheus-Storage](http://prom.local:30003)

## Postgres HA Cluster

**Questions**:

- How do I automatically deploy a new PostgreSQL instance?
- How do I failover a PostgreSQL pod to another availability zone if my PostgreSQL instance goes down?
- How do I resize my PostgreSQL volume if I am running out of space?
- How do I snapshot and backup PostgreSQL for disaster recovery?
- How do I test upgrades?
- Can I take my PostgreSQL deployment and run it in any environment if needed?

**Achieving HA with PostgreSQL**:

- [Zalando Operator](https://github.com/zalando/postgres-operator)
