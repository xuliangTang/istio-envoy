package controllers

import (
	"context"
	"fmt"
	"istio-envoy/mygateway/bootstrap"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	IngressClassName = "myenvoy"
)

type IngressController struct {
	client.Client
	E record.EventRecorder
}

func NewIngressController(e record.EventRecorder) *IngressController {
	return &IngressController{E: e}
}

func (r *IngressController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	ing := &networkingv1.Ingress{}
	err := r.Get(ctx, req.NamespacedName, ing)
	if err != nil {
		return reconcile.Result{}, err
	}

	if ing.Spec.IngressClassName != nil && *ing.Spec.IngressClassName == IngressClassName {
		fmt.Println("接收到ingress资源:", ing.Name)
		bootstrap.ApplyIngress(ing)
	}

	return reconcile.Result{}, nil
}

// OnDelete 监听删除ingress资源
func (r *IngressController) OnDelete(event event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface) {
	if ing, ok := event.Object.(*networkingv1.Ingress); ok && *ing.Spec.IngressClassName == IngressClassName {
		log.Println("删除了ing：", ing.Name)
		bootstrap.RemoveIngress(ing)
	}
}

func (r *IngressController) InjectClient(c client.Client) error {
	r.Client = c
	return nil
}
