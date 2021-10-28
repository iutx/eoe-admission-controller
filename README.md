# eoe admission controller

## How to build image ?

```shell
make build-image
```

## How to build charts ?

```shell
helm package charts/eoe-admission -d ./packages
```

## How to install?

```shell
 helm install eoe-admission packages/eoe-admission-{version}.tgz -n {namespace} --create-namespace 
```