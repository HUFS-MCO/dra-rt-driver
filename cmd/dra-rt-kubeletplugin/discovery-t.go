package main

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	// v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	// metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
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
	nodes, _ := c.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	fmt.Println("nodeNames:", nodes)
	node, err := c.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})

	// // Create metrics clientset for accessing metrics API
	// metricsClient, err := metricsclientset.NewForConfig(cfg)
	// if err != nil {
	// 	fmt.Printf("error creating metrics client: %v\n", err)
	// }

	// // Get node metrics
	// nodeMetrics, err := getNodeMetrics(metricsClient, nodeName)
	// if err != nil {
	// 	fmt.Printf("error getting node metrics: %v\n", err)
	// }

	// fmt.Printf("Node %s Metrics:\n", nodeName)
	// fmt.Printf("CPU Usage: %v\n", nodeMetrics.Usage.Cpu().String())
	// fmt.Printf("Memory Usage: %v\n", nodeMetrics.Usage.Memory().String())

	cpuset := node.Status.Capacity.Cpu().Value()
	fmt.Println("cpuset:", cpuset)
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

// // Get node metrics from metrics API
// func getNodeMetrics(metricsClient *metricsclientset.Clientset, nodeName string) (*v1beta1.NodeMetrics, error) {
// 	nodeMetrics, err := metricsClient.MetricsV1beta1().NodeMetricses().Get(context.TODO(), nodeName, metav1.GetOptions{})
// 	if err != nil {
// 		return nil, fmt.Errorf("error fetching node metrics: %v", err)
// 	}
// 	return nodeMetrics, nil
// }
