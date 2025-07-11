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
| affinity | object | `{}` |  |
| autoscaling.enabled | bool | `false` | Enable replica autoscaling settings |
| autoscaling.maxReplicas | int | `11` | Maximum replicas for the pod autoscaling |
| autoscaling.minReplicas | int | `1` | Minimum replicas for the pod autoscaling |
| autoscaling.targetCPU | string | `"80"` | Percentage of CPU to consider when autoscaling |
| autoscaling.targetMemory | string | `""` | Percentage of Memory to consider when autoscaling |
| containerSecurityContext.allowPrivilegeEscalation | bool | `false` |  |
| containerSecurityContext.capabilities.drop[0] | string | `"ALL"` |  |
| containerSecurityContext.readOnlyRootFilesystem | bool | `true` |  |
| containerSecurityContext.runAsNonRoot | bool | `true` |  |
| containerSecurityContext.runAsUser | int | `1000` |  |
| containerSecurityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| extraEnvVars | list | `[]` |  |
| extraLabels | object | `{}` |  |
| extraObjects | list | `[]` | Additional objects to deploy |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ghcr.io/giuliocalzolari/ecr-proxy-go"` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.enabled | bool | `false` |  |
| ingress.existingSecret | string | `""` |  |
| ingress.hostname | string | `"chart-example.local"` |  |
| ingress.ingressClassName | string | `"nginx"` |  |
| ingress.path | string | `"/"` |  |
| ingress.pathType | string | `"Prefix"` |  |
| ingress.tls | bool | `true` |  |
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
| service.annotations | object | `{}` |  |
| service.name | string | `"ecr-proxy"` |  |
| service.port | int | `5000` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `false` |  |
| serviceAccount.name | string | `"ecr-proxy"` |  |
| tolerations | list | `[]` |  |
| topologySpreadConstraints | list | `[]` |  |

