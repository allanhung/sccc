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
### conf/app2.properties
```yaml
outputDir=/app/data
apiKey=default_key
```
### conf/app3.properties
```yaml
namespace:
  dev:
    apiKey=dev_key
version:
  1.0.6:
    apiKey=1.0.6_key
    newProperty=1.0.6_newProperty
```
### conf/app1.yaml
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
### conf/app2.yaml
```yaml
replicaCount: 2
image:
  repository: nginx
  tag: stable
  pullPolicy: IfNotPresent
imagePullSecrets: []
service:
  type: ClusterIP
  port: 8080
```  
### conf/app3.yaml
```yaml
namespace:
  prod:
    replicaCount: 4
    service:
      type: NLB
  dev:
    replicaCount: 1
    service:
### resources/myres1
```
### resources/myres1
```
# test
123
```

## on client side
```bash
sccc get --help                                                                                                                   ✔  14:48:16  100% 🔋
get config from spring cloud config server
For example:

sccc get -u http://localhost:8888 -a app -n dev -v 1.0.6 -b master \
         -c conf/app1.properties=/app/application1.properties \
         -c conf/app1.yaml=/app/application1.yaml \
         -r resources/myres1=/app/myres1.res \
         -r resources/myres2=/app/myres2.res
or
sccc get -u http://localhost:8888 -a app -n dev -v 1.0.6 -b master \
         -c conf/app2.properties:conf/app3.properties=/app/application1.properties \
         -c conf/app2.yaml:conf/app3.yaml=/app/application1.yaml \
         -r resources/myres1=/app/myres1.res \
         -r resources/myres2=/app/myres2.res

Usage:
  sccc get [flags]

Flags:
  -a, --application string         application default: application (default "application")
  -b, --branch string              git branch default: master (default "master")
  -c, --configfile configFiles     config file example: conf/app.conf=/etc/application.propertiess (can specify multiple) (default [])
  -h, --help                       help for get
  -n, --namespace string           kubernetes namespace
  -r, --resourcefile configFiles   resource file example: resources/myres=/app/app.res (can specify multiple) (default [])
  -u, --uri string                 spring cloud config server uri (default "http://localhost:8888")
  -v, --version string             application version

Global Flags:
      --config string   config file (default is $HOME/.sccc.yaml)
```
```bash
sccc get -u http://localhost:8888 -a app -n dev -v 1.0.6 -b master \
         -c conf/app1.properties=/app/application1.properties \
         -c conf/app1.yaml=/app/application1.yaml \
         -r resources/myres1=/app/myres1.res
or
sccc get -u http://localhost:8888 -a app -n dev -v 1.0.6 -b master \
         -c conf/app2.properties:conf/app3.properties=/app/application1.properties \
         -c conf/app2.yaml:conf/app3.yaml=/app/application1.yaml \
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
