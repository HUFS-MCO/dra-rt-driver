package main

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func enumerateCpusets() (AllocatableRtCpus, error) {

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
		return nil, fmt.Errorf("error building in-cluster config: %v", err)
	}
	fmt.Println("cluster config is ready")

	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error building kubernetes client: %v", err)
	}
	fmt.Println("kubernetes client is ready")

	pod_list, e := c.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if e != nil {
		return nil, fmt.Errorf("error listing pods: %v", e)
	}

	for i, p := range pod_list.Items {
		fmt.Printf("Pod %d: %s\n", i, p.Name)

	}

	// List nodes
	// nodes, err := c.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	// if err != nil {
	// 	panic(err.Error())
	// }

	nodeName := os.Getenv("NODE_NAME")
	node, err := c.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	cpuset := node.Status.Capacity.Cpu().Value()
	ids := make([]int, cpuset)

	alldevices := make(AllocatableRtCpus)
	for _, id := range ids {
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
