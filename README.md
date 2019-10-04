# SCCC
Spring cloud-config-client written in Go.

It fetches file directly from a Spring cloud-config-server and merge the config by flag.

The goal is fetch config as init container in kubernetes and support config by kubernetes metadata and labels.

# Example
## on server side
### conf/app1.properties
```yaml
default:
  outputDir=/app/data
  apiKey=default_key
namespace:
  dev:
    apiKey=dev_key
version:
  1.0.6:
    apiKey=1.0.6_key
    newProperty=1.0.6_newProperty
```
### conf/app2.yaml
```yaml
default:
  replicaCount: 2
  image:
    repository: nginx
    tag: stable
    pullPolicy: IfNotPresent
  imagePullSecrets: []
  service:
    type: ClusterIP
    port: 8080
namespace:
  prod:
    replicaCount: 4
    service:
      type: NLB
  dev:
    replicaCount: 1
    service:
      type: DNS
version:
  1.0.6:
    service:
      type: pod
    newproperty: test
```
### resources/myres1
```
# test
123
```

## on client side
```bash
sccc get -u http://localhost:8888 -a app -n dev -v 1.0.6 -b master \
         -c conf/app1.properties=/app/application1.properties \
         -c conf/app2.yaml=/app/application2.yaml \
         -r resources/myres1=/app/myres1.res
```
### result
### /app/application1.properties
```
outputDir=/app/data
apiKey=1.0.6_key
newProperty=1.0.6_newProperty
```
### /app/application2.yaml
```
replicaCount: 1
image:
  repository: nginx
  tag: stable
  pullPolicy: IfNotPresent
imagePullSecrets: []
service:
  type: pod
  port: 8080
newproperty: test
```
### /app/myres1.res
```
# test
123
```

# Roadmap

- [] Dockerfile
- [] Config Compare

Here's a rough outline on what is to come (subject to change):

### v0.1

- [x] Support Config Merge
