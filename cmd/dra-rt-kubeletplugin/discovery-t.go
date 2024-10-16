package main

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	// v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1alpha2"
)

func enumerateCpusets() (AllocatableRtCpus, error) {

	var cfg *rest.Config
	var err error

	// Use in-cluster configuration if running inside a Kubernetes pod
	cfg, err = rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("error building in-cluster config: %v", err)
	}
	fmt.Println("cluster config is ready")

	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error building kubernetes client: %v", err)
	}
	fmt.Println("kubernetes client is ready")

	// pod_list, e := c.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	// if e != nil {
	// 	return nil, fmt.Errorf("error listing pods: %v", e)
	// }

	// for i, p := range pod_list.Items {
	// 	fmt.Printf("Pod %d: %s\n", i, p.Name)

	// }

	nodeName := os.Getenv("NODE_NAME")
	nodes, _ := c.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	fmt.Println("nodeNames:", nodes)
	node, err := c.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})

	cpuset := node.Status.Capacity.Cpu().Value()
	fmt.Println("cpuset:", cpuset)

	alldevices := make(AllocatableRtCpus)
	for id := 0; id < int(cpuset); id++ {
		deviceInfo := &AllocatableCpusetInfo{
			RtCpuInfo: &RtCpuInfo{
				id:   id,
				util: 0,
			},
		}
		alldevices[id] = deviceInfo
	}
	return alldevices, nil
}
