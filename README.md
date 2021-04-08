# Network Monitoring Operator

A Kubernetes operator written in Go and using the client-go Kubernetes library. It monitors the status of the network by means of the "iwconfig" Linux tool. If the connectivity is not good enough, it will interact with the Kubernetes cluster running on the node it is executed onto and changes the Selector field of the service called "yolo-service" belonging to the namespace "yolo" from "remote" to "local". The exact opposite happens when the network connectivity will become good again. This is an implementation meant to run outside of the cluster.
