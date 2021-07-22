# dok8cert

![pipeline](https://github.com/adamkgray/dok8cert/actions/workflows/go.yml/badge.svg)


Update the go kubernetes client with the custom DigitalOcean TLS certificate

## About

When connecting to the kubernetes API on a DigitalOcean kubernetes cluster, it is possible for the requests to fail with a certificate error. This is because DigitalOcean provides a custom TLS certificate with kubernetes credentials, and you have to use that one for the go client.

As seen in [this SO thread](https://stackoverflow.com/questions/65042279/python-kubernetes-client-requests-fail-with-unable-to-get-local-issuer-certific), the error can also happen for the python client.

I have not been able to find the documentation explaining why you must use this custom cert, but this package at least solves the problem and lets you continue coding ^.^

## How to use it

This package exposes a single method `Update()`

Example:

```go
// instantiate an in-cluster client as usual
config, err := rest.InClusterConfig()
if err != nil {
    // ...
}

// update the cert
ok, err := dok8cert.Update("clusterId", "accessToken", config)
if err != nil {
    // ...
}

// continue doing the thing
clientset, err := kubernetes.NewForConfig(config)
```
