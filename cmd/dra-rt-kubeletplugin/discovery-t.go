package main

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func enumerateCpusets() error {

	// Define the kubeconfig path
	// var kubeconfig *string
	// if home := homedir.HomeDir(); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join("$HOME", ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }
	// flag.Parse()

	// cfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	return fmt.Errorf("error building kubeconfig: %v", err)
	// }

	var cfg *rest.Config
	var err error

	// Use in-cluster configuration if running inside a Kubernetes pod
	cfg, err = rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("error building in-cluster config: %v", err)
	}
	fmt.Println("cluster config is ready")

	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building kubernetes client: %v", err)
	}
	fmt.Println("kubernetes client is ready")

	pod_list, e := c.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if e != nil {
		return fmt.Errorf("error listing pods: %v", e)
	}

	for i, p := range pod_list.Items {
		fmt.Printf("Pod %d: %s\n", i, p.Name)
	}

	// List nodes
	nodes, err := c.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, node := range nodes.Items {
		fmt.Printf("Node Name: %s\n", node.Name)

		// Get the CPU capacity of the node
		cpuCapacity := node.Status.Capacity["cpu"]
		fmt.Printf("CPU Capacity: %s\n", cpuCapacity.String())

		// Optionally, get allocatable CPU (available for pods)
		cpuAllocatable := node.Status.Allocatable["cpu"]
		fmt.Printf("CPU Allocatable: %s\n", cpuAllocatable.String())

		// You could also check for annotations or labels if custom ones exist
		if cpuset, ok := node.Annotations["cpuset"]; ok {
			fmt.Printf("Cpuset annotation: %s\n", cpuset)
		}
		cpusets := node.Status.Capacity.Cpu()
		fmt.Println("Cpusets:", cpusets)
	}
	return nil
}
