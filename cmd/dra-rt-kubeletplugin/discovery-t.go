package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func enumerateCpusets() error {

	// Define the kubeconfig path
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return fmt.Errorf("error building kubeconfig: %v", err)
	}

	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("error building kubernetes client: %v", err)
	}
	pod_list, e := c.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if e != nil {
		return fmt.Errorf("error listing pods: %v", e)
	}

	for i, p := range pod_list.Items {
		fmt.Printf("Pod %d: %s\n", i, p.Name)
		// p.Spec.NodeName
		// (SetAnnotation("period", period_as_a_string) or similar)
	}
	return nil
}
