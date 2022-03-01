 # kubernetes-podlogger

Two seperate golang app (client-server) to monitor kubernetes pods inside their namespaces and communicate with each other when pod events occur

## Technologies

The project has been created using these technologies:

* **golang** as programming language
* **Kubernetes** as container-orchestration system
* **kubectl** as command-line tool to interact kubernetes
* **client-go** for talking to a kubernetes cluster
* **ws** tiny websocket library for Golang

## Setup & Installtion

**Install:**

1. `golang` from <a href="https://golang.org/dl/">here</a>
2. `Docker` from <a href="https://docs.docker.com/get-docker/">here</a> then enable kubernetes
3. `kubectl` from <a href="https://kubernetes.io/docs/tasks/tools/">here</a>
4. `helm` from <a href="https://helm.sh/docs/intro/install/">here</a>

## Running The App

Create namespaces for client and server:
-  `ns1` for server -> `kubectl create namespace ns1`
-  `ns2` for client -> `kubectl create namespace ns2`

Execute these commands to install the helm charts:

```
helm install podloggerserver ./podlogger-server/helm
helm install podloggerclient ./podlogger-client/helm
```

Check your pod's logs:

```
kubectl logs -f -n <namespace> -l "app=<app-name>"
```
