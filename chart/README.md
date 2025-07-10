# ecr-proxy-helm

A proxy for ECR that handles authentication and caching

## How to install this chart

A simple install with default values:

```console
helm install my-release oci://ghcr.io/giuliocalzolari/ecr-proxy-helm
```

To install with some set values:

```console
helm install my-release oci://ghcr.io/giuliocalzolari/ecr-proxy-helm --set values_key1=value1 --set values_key2=value2
```

To install with custom values file:

```console
helm install my-release oci://ghcr.io/giuliocalzolari/ecr-proxy-helm -f values.yaml
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| extraEnvVars | list | `[]` |  |
| extraLabels | object | `{}` |  |
| fullnameOverride | string | `""` |  |
| image.args | list | `[]` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ghcr.io/giuliocalzolari/ecr-proxy-go"` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts[0].host | string | `"chart-example.local"` |  |
| ingress.hosts[0].paths[0] | string | `"/"` |  |
| ingress.ingressClassName | string | `nil` |  |
| ingress.tls | list | `[]` |  |
| ipWhitelist | list | `[]` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podSecurityContext.fsGroup | int | `1000` |  |
| podSecurityContext.fsGroupChangePolicy | string | `"Always"` |  |
| podSecurityContext.runAsNonRoot | bool | `true` |  |
| podSecurityContext.runAsUser | int | `1000` |  |
| podSecurityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| priorityClassName | string | `""` |  |
| replicaCount | int | `1` |  |
| resources.limits.cpu | int | `1` |  |
| resources.limits.memory | string | `"1Gi"` |  |
| resources.requests.cpu | string | `"150m"` |  |
| resources.requests.memory | string | `"256Mi"` |  |
| securityContext.allowPrivilegeEscalation | bool | `false` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.runAsUser | int | `1000` |  |
| securityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| service.annotations | object | `{}` |  |
| service.name | string | `"ecr-proxy"` |  |
| service.port | int | `5000` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `false` |  |
| serviceAccount.name | string | `"ecr-proxy"` |  |
| tolerations | list | `[]` |  |

