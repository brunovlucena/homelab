**Helm Operator**

Operator from scratch

```sh
$ operator-sdk new myapp-helm-operator --api-version=example.com/v1alpha1 --kind=AppServiceHelm --type=helm
```

Operator using `deployments/chart`

```sh
$ operator-sdk new myapp-helm-operator --api-version=example.com/v1alpha1 --kind=AppServiceHelm --type=helm --helm-chart-repo deployments/charts
```
