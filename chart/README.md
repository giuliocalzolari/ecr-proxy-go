# ecr-proxy

A proxy for ECR that handles authentication and caching

## How to install this chart

A simple install with default values:

```console
helm install my-release oci://ghcr.io/giuliocalzolari/ecr-proxy
```

To install with some set values:

```console
helm install my-release oci://ghcr.io/giuliocalzolari/ecr-proxy --set values_key1=value1 --set values_key2=value2
```

To install with custom values file:

```console
helm install my-release oci://ghcr.io/giuliocalzolari/ecr-proxy -f values.yaml
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
|  | string | `nil` |  |

