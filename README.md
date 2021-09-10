# node-dns

`node-dns` is a simple local DNS server that can be deployed via kubernetes. It is specially built for using with [KubeEdge](github.com/kubeedge/kubeedge) nodes.
`node-dns` only works for the very node only where it is running. You must ensure that your pods that want to communicate whith each other run on the exact same node. 

# Why?

When using KubeEdge, it is common for two pods running on the same node to need to communicate with each other.
Due to the fact that containers running on the edge using KubeEdge cannot address each other with their container host names, they can't easily find them other than the using the IP address. However the IP address of the containers might not be known to other containers or change if a container dies and gets rescheduled. `node-dns` solves the problem of resolving a name with the container IP addresses in a very light weight way directly on the edge node.

# How?

`node-dns` gets it's information on how to resolve which host to its IP address by configuring and using a `feed`. 
A `feed` is a source for information to and could be implemented as e.g. 
* the k8s api server
* scanning currently containers using a docker client.
`node-dns` updates its `feed` every minute. This means that in the worst case the DNS resolution is possible after 60 seconds max.

## Currently supported feeds

It might be possible to introduce more feeds, however currently supported is the k8s API feed only.

### k8s API feed
The k8s API feed connects directly to the k8s API server (see configuration file), reads the podsList and looks at the pods labels. All pods with the label `node-dns.host=<value>` will be handeled by `node-dns`. 

# Examples
See the `examples/` directory for example manifest files on hwo to use the needed label.

# Configuration

A valid configuration can look like this:
```yaml
listeninterface: docker0
listenport: 53
feed:
  k8sapi:
    enabled: true
    uri: http://127.0.0.1:10550
    insecuretls: true
    token: ""
```