# dok8cert

A wrapper for digital ocean kubernetes client certificate management

## About

When connecting to the kubernetes API on a DigitalOcean kubernetes cluster, it is possible for the requests to fail with a certificate error. This is because DigitalOcean provides a custom TLS certificate with kubernetes credentials, and you have to use that one for the go client.

As seen in [this SO thread](https://stackoverflow.com/questions/65042279/python-kubernetes-client-requests-fail-with-unable-to-get-local-issuer-certific), the error can also happen for the python client.

I have not been able to find the documentation explaining why you must use this custom cert, but this package at least solves the problem and lets you continue coding ^.^

## How to use it

This package provides two methods: `Get()` and `Set()`

1. `Get()` takes your cluster id and access token and returns the custom certificate as a slice of bytes

2. `Set()` takes the cert slice and your rest client sets the new cert in the correct place

Example:

```go
// instantiate an in-cluster client as usual
config, err := rest.InClusterConfig()
if err != nil {
    // ...
}

// get the cert
cert, err := dok8cert.Get("clusterId", "accessToken")
if err != nil {
    // ...
}

// put the cert in your client
dok8cert.Set(cert, config)

// continue doing the thing
clientset, err := kubernetes.NewForConfig(config)
```
