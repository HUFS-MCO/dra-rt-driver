package controller

import (
	"context"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type Controller struct {
	driver        Driver
	clientset     kubernetes.Interface
	queue         workqueue.RateLimitingInterface
	claimInformer cache.SharedIndexInformer
	apiGroup      string
}

func New(ctx context.Context, apiGroup string, driver Driver, clientset kubernetes.Interface, factory informers.SharedInformerFactory) *Controller {
	claimInformer := factory.Resource().V1().ResourceClaims().Informer()

	c := &Controller{
		driver:        driver,
		clientset:     clientset,
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "dra-controller"),
		claimInformer: claimInformer,
		apiGroup:      apiGroup,
	}

	claimInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.enqueue,
		UpdateFunc: func(old, new interface{}) { c.enqueue(new) },
		DeleteFunc: func(obj interface{}) { c.enqueue(obj) },
	})

	return c
}

func (c *Controller) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("Failed to get key: %v", err)
		return
	}
	c.queue.Add(key)
}

func (c *Controller) Run(workers int) {
	defer c.queue.ShutDown()

	klog.Info("Starting DRA controller")
	defer klog.Info("Shutting down DRA controller")

	for i := 0; i < workers; i++ {
		go c.worker()
	}

	<-make(chan struct{}) // Block forever
}

func (c *Controller) worker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	key, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(key)

	// 실제 처리 로직은 기존 드라이버가 처리하므로 여기서는 기본적인 큐 관리만
	klog.V(4).Infof("Processing claim: %v", key)

	c.queue.Forget(key)
	return true
}
