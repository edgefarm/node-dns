# Examples

Here you can find some examples to get the idea of the correct placement and usage of `node-dns`. 

Per default all examples are scheduled on all nodes which have the following labels set by default. These labels are set by default for KubeEdge nodes.
* `node-role.kubernetes.io/edge=` 
* `node-role.kubernetes.io/agent=`

Modify tolerations if needed any.

:exclamation: **Note:** :exclamation:
For the `deployment.yaml` example, you need to modify the value for the label `kubernetes.io/hostname` to match the node you want to schedule this on. This is needed because the deployment does not guarantee that the pods are scheduled on the same node. However this is mandatory for `node-dns` to work properly. Otherwise it might or might not work as you'd expect.