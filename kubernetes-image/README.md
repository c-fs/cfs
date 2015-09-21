# Cloud Native Deployments of cfs using Kubernetes

The following document describes the development of a *cloud native* cfs deployment on Kubernetes. Cloud native deployment indicates application understands that it is running within a cluster manager, and uses this cluster management infrastructure to help implement the application[1].

## Preparation

You should have a Kubernetes cluster running and `kubectl` command line ready. Read [getting started](https://github.com/kubernetes/kubernetes/blob/master/docs/getting-started-guides) for upstream installation instructions.

You should know basics about Kubernetes concepts, including pods, service and replication controller. [Basics tutorials](http://kubernetes.io/v1.0/basicstutorials.html) explain them well.

## Create Single cfs Instance

Create a replication controller for cfs instances. Its default replica count is 1, so it creates single cfs instance in the cluster.

```
kubectl create -f cfs-controller.yaml
```

Check replication controller's status and pods' status:

```
kubectl describe rc cfs
kubectl describe pods --selector=name=cfs
```

Check cfs server in the pod is ready:
[TODO: read and write from one cfs server]

```
kubectl logs cfs-xxxxx
```

## Scale up

Scale up cfs instances to 2:

```
kubectl scale rc cfs --replicas=2
```

[TODO: read and write from all cfs servers using cfs nameserver]

## Clean up

```
kubectl delete rc cfs
```

[1]: https://github.com/kubernetes/kubernetes/tree/master/examples/cassandra
