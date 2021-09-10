# Examples

Here you can find some examples to get the idea of the correct placement and usage of `node-dns`.

Per default all examples are scheduled on all nodes which have the following labels set by default. These labels are set by default for KubeEdge nodes.
* `node-role.kubernetes.io/edge=`
* `node-role.kubernetes.io/agent=`

Modify tolerations if needed any.

# Notes

:exclamation:  **`deployment.yaml`** and **`pods.yaml`** :exclamation:

For the `deployment.yaml` example, you need to modify the value for the label `kubernetes.io/hostname` to match the node you want to schedule this on.
For `pods.yaml` example you have to modify the `nodeName` property to whatever node you want to schedule both pods.
This is needed because otherwise the scheduler does not guarantee that the pods are scheduled on the same node. However this is mandatory for `node-dns` to work properly. Otherwise it might or might not work as you'd expect.

To test if it works look at the logs of your curl pod. It might take a little while before the update cycle of the DNS server collects all new information and make the name resolution working.

# Usage

This shows an example run of the pods example. The other examples are used just the same.

```sh
$ kubectl apply -f pods.yaml
pod/curl-pod created
pod/nginx-pod created
$ kubectl get pods -o wide
NAME        READY   STATUS    RESTARTS   AGE     IP           NODE           NOMINATED NODE   READINESS GATES
curl-pod    1/1     Running   0          2m34s   172.17.0.5   mynode         <none>           <none>
nginx-pod   1/1     Running   0          2m33s   172.17.0.6   mynode         <none>           <none>
# give the dns some time to update its cache
$ kubectl logs curl-pod
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0curl: (6) Could not resolve host: nginx.nginx-pod
Fri Sep 10 09:53:27 UTC 2021
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   615  100   615    0     0   182k      0 --:--:-- --:--:-- --:--:--  300k
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
html { color-scheme: light dark; }
body { width: 35em; margin: 0 auto;
font-family: Tahoma, Verdana, Arial, sans-serif; }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
Fri Sep 10 09:53:28 UTC 2021
```
