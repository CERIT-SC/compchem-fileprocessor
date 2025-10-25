# compchem

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.16.0](https://img.shields.io/badge/AppVersion-1.16.0-informational?style=flat-square)

A Helm chart for the CompcChem component, this chart has not been tested and is not complete!

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | compchem-minio(minio) | 17.0.21 |
| https://charts.bitnami.com/bitnami | compchemPostgres(postgresql) | 16.7.27 |
| https://charts.bitnami.com/bitnami | compchem-rabbitmq(rabbitmq) | 16.0.14 |
| https://charts.bitnami.com/bitnami | compchem-redis(redis) | 22.0.7 |
| https://opensearch-project.github.io/helm-charts | compchem-opensearch(opensearch) | 3.2.1 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| compchem-minio.auth.rootPassword | string | `"minio-password123"` |  |
| compchem-minio.auth.rootUser | string | `"compchem"` |  |
| compchem-minio.defaultBuckets | string | `"compchem-bucket"` |  |
| compchem-minio.enabled | bool | `true` |  |
| compchem-minio.metrics.enabled | bool | `false` |  |
| compchem-minio.persistence.enabled | bool | `true` |  |
| compchem-minio.persistence.size | string | `"20Gi"` |  |
| compchem-minio.resources.limits.cpu | string | `"500m"` |  |
| compchem-minio.resources.limits.memory | string | `"512Mi"` |  |
| compchem-minio.resources.requests.cpu | string | `"250m"` |  |
| compchem-minio.resources.requests.memory | string | `"256Mi"` |  |
| compchem-minio.service.ports.api | int | `9000` |  |
| compchem-minio.service.ports.console | int | `9001` |  |
| compchem-minio.service.type | string | `"ClusterIP"` |  |
| compchem-opensearch.config."opensearch.yml" | string | `"cluster.name: compchem-cluster\nnetwork.host: 0.0.0.0\nbootstrap.memory_lock: true\ndiscovery.type: single-node\nplugins.security.disabled: true\nplugins.security.ssl.http.enabled: false\nplugins.security.ssl.transport.enabled: false\n"` |  |
| compchem-opensearch.enabled | bool | `true` |  |
| compchem-opensearch.masterService | string | `"compchem-opensearch"` |  |
| compchem-opensearch.opensearchJavaOpts | string | `"-Xmx512m -Xms512m"` |  |
| compchem-opensearch.persistence.enabled | bool | `true` |  |
| compchem-opensearch.persistence.size | string | `"10Gi"` |  |
| compchem-opensearch.resources.limits.cpu | string | `"1000m"` |  |
| compchem-opensearch.resources.limits.memory | string | `"1Gi"` |  |
| compchem-opensearch.resources.requests.cpu | string | `"500m"` |  |
| compchem-opensearch.resources.requests.memory | string | `"512Mi"` |  |
| compchem-opensearch.service.ports.http | int | `9200` |  |
| compchem-opensearch.service.ports.transport | int | `9300` |  |
| compchem-opensearch.service.type | string | `"ClusterIP"` |  |
| compchem-opensearch.singleNode | bool | `true` |  |
| compchem-rabbitmq.auth.erlangCookie | string | `"secretcookie123"` |  |
| compchem-rabbitmq.auth.password | string | `"rabbitmq-password123"` |  |
| compchem-rabbitmq.auth.username | string | `"compchem"` |  |
| compchem-rabbitmq.enabled | bool | `true` |  |
| compchem-rabbitmq.metrics.enabled | bool | `false` |  |
| compchem-rabbitmq.persistence.enabled | bool | `true` |  |
| compchem-rabbitmq.persistence.size | string | `"8Gi"` |  |
| compchem-rabbitmq.resources.limits.cpu | string | `"500m"` |  |
| compchem-rabbitmq.resources.limits.memory | string | `"512Mi"` |  |
| compchem-rabbitmq.resources.requests.cpu | string | `"250m"` |  |
| compchem-rabbitmq.resources.requests.memory | string | `"256Mi"` |  |
| compchem-rabbitmq.service.ports.amqp | int | `5672` |  |
| compchem-rabbitmq.service.ports.manager | int | `15672` |  |
| compchem-redis.auth.enabled | bool | `false` |  |
| compchem-redis.enabled | bool | `true` |  |
| compchem-redis.master.persistence.enabled | bool | `false` |  |
| compchem-redis.master.resources.limits.cpu | string | `"250m"` |  |
| compchem-redis.master.resources.limits.memory | string | `"256Mi"` |  |
| compchem-redis.master.resources.requests.cpu | string | `"100m"` |  |
| compchem-redis.master.resources.requests.memory | string | `"128Mi"` |  |
| compchem-redis.master.service.ports.redis | int | `6379` |  |
| compchem-redis.metrics.enabled | bool | `false` |  |
| compchem-redis.replica.replicaCount | int | `0` |  |
| compchemPostgres.auth.database | string | `"compchem"` |  |
| compchemPostgres.auth.password | string | `"compchem-password123"` |  |
| compchemPostgres.auth.postgresPassword | string | `"compchem-postgres123"` |  |
| compchemPostgres.auth.username | string | `"compchem"` |  |
| compchemPostgres.enabled | bool | `true` |  |
| compchemPostgres.metrics.enabled | bool | `false` |  |
| compchemPostgres.primary.persistence.enabled | bool | `true` |  |
| compchemPostgres.primary.persistence.size | string | `"20Gi"` |  |
| compchemPostgres.primary.resources.limits.cpu | string | `"1000m"` |  |
| compchemPostgres.primary.resources.limits.memory | string | `"1Gi"` |  |
| compchemPostgres.primary.resources.requests.cpu | string | `"500m"` |  |
| compchemPostgres.primary.resources.requests.memory | string | `"512Mi"` |  |
| compchemPostgres.primary.service.ports.postgresql | int | `5432` |  |
| database.host | string | `"external-postgres.example.com"` |  |
| database.name | string | `"compchem"` |  |
| database.password | string | `"external-password"` |  |
| database.port | int | `5432` |  |
| database.user | string | `"compchem"` |  |
| enabled | bool | `false` |  |
| flaskEnv | string | `"production"` |  |
| image.pullPolicy | string | `"Always"` |  |
| image.repository | string | `"your-registry/compchem"` |  |
| image.tag | string | `"latest"` |  |
| ingress.annotations."nginx.ingress.kubernetes.io/backend-protocol" | string | `"HTTPS"` |  |
| ingress.annotations."nginx.ingress.kubernetes.io/ssl-redirect" | string | `"true"` |  |
| ingress.className | string | `""` |  |
| ingress.enabled | bool | `true` |  |
| ingress.hosts[0].host | string | `"compchem.local"` |  |
| ingress.hosts[0].paths[0].path | string | `"/"` |  |
| ingress.hosts[0].paths[0].pathType | string | `"Prefix"` |  |
| ingress.tls[0].hosts[0] | string | `"compchem.local"` |  |
| ingress.tls[0].secretName | string | `"compchem-tls"` |  |
| instancePath | string | `"/invenio/instance"` |  |
| nodeSelector | object | `{}` |  |
| opensearch.clusterPort | int | `9300` |  |
| opensearch.host | string | `"external-opensearch.example.com"` |  |
| opensearch.port | int | `9200` |  |
| persistence.accessMode | string | `"ReadWriteOnce"` |  |
| persistence.enabled | bool | `true` |  |
| persistence.size | string | `"10Gi"` |  |
| rabbitmq.adminPort | int | `15672` |  |
| rabbitmq.host | string | `"external-rabbitmq.example.com"` |  |
| rabbitmq.password | string | `"external-rabbitmq-password"` |  |
| rabbitmq.port | int | `5672` |  |
| rabbitmq.user | string | `"compchem"` |  |
| redis.host | string | `"external-redis.example.com"` |  |
| redis.port | int | `6379` |  |
| replicaCount | int | `1` |  |
| resources.limits.cpu | string | `"2000m"` |  |
| resources.limits.memory | string | `"2Gi"` |  |
| resources.requests.cpu | string | `"500m"` |  |
| resources.requests.memory | string | `"1Gi"` |  |
| s3.accessKey | string | `"external-access-key"` |  |
| s3.bucket | string | `"compchem-bucket"` |  |
| s3.consolePort | int | `9001` |  |
| s3.endpointUrl | string | `"https://external-s3.example.com"` |  |
| s3.host | string | `"external-s3.example.com"` |  |
| s3.port | int | `9000` |  |
| s3.secretKey | string | `"external-secret-key"` |  |
| secretKey | string | `"your-super-secret-key-change-this-in-production"` |  |
| service.httpPort | int | `5000` |  |
| service.httpsPort | int | `8443` |  |
| service.statsPort | int | `6969` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| staticFolder | string | `"/invenio/instance/static"` |  |
| theme | string | `"compchem"` |  |
| tls.cert | string | `""` |  |
| tls.key | string | `""` |  |
| tolerations | list | `[]` |  |
| wsgi.processes | int | `2` |  |
| wsgi.threads | int | `4` |  |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
