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

## Local develop

```shell
kubectl get secret eoe-admission  -o jsonpath='{.data.ca}'| base64 -d > ca
kubectl get secret eoe-admission  -o jsonpath='{.data.tls\.crt}'| base64 -d > tls.crt
kubectl get secret eoe-admission -o jsonpath='{.data.tls\.key}'| base64 -d > tls.key
```



