package k8sconfig

import (
	"istio-envoy/mygateway/bootstrap"
	"istio-envoy/mygateway/operator/controllers"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var SchemeBuilder = &Builder{}

type Builder struct {
	runtime.SchemeBuilder
}

func (this *Builder) AddScheme(scheme *runtime.Scheme) error {
	return this.AddToScheme(scheme)
}

func InitManager() {
	logf.SetLogger(zap.New())
	mgr, err := manager.New(K8sRestConfig(), manager.Options{
		Logger: logf.Log.WithName("gr"),
	})
	if err != nil {
		log.Fatal("创建管理器失败:", err.Error())
	}

	// Schema定义了资源序列化和反序列化的方法以及资源类型和版本的对应关系
	if err = SchemeBuilder.AddToScheme(mgr.GetScheme()); err != nil {
		mgr.GetLogger().Error(err, "unable add schema")
		os.Exit(1)
	}

	// 初始化控制器对象
	ingController := controllers.NewIngressController(mgr.GetEventRecorderFor("gr"))

	// 构建controller
	err = builder.ControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}).
		Complete(ingController)
	if err != nil {
		mgr.GetLogger().Error(err, "unable to create manager")
		os.Exit(1)
	}

	// 启动控制面server
	if err := mgr.Add(bootstrap.NewGatewayBooter()); err != nil {
		mgr.GetLogger().Error(err, "unable to create gateway server")
		os.Exit(1)
	}

	if err = mgr.Start(signals.SetupSignalHandler()); err != nil {
		mgr.GetLogger().Error(err, "unable to start manager")
	}

}
