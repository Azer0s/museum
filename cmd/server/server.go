package server

import (
	"fmt"
	docker "github.com/docker/docker/client"
	etcd "go.etcd.io/etcd/client/v3"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/config"
	"museum/http/api"
	"museum/http/exhibit"
	"museum/http/health"
	"museum/http/router"
	"museum/ioc"
	"museum/observability"
	"museum/persistence"
	"museum/service"
	"museum/service/impl"
	"museum/util"
	"net/http"
)

func Run() {
	c := ioc.NewContainer()

	// register logger
	ioc.RegisterSingleton[*zap.SugaredLogger](c, observability.NewLogger)

	// register config
	ioc.RegisterSingleton[config.Config](c, config.NewEnvConfig)

	// register docker
	ioc.RegisterSingleton[*docker.Client](c, service.NewDockerClient)

	// register jaeger
	ioc.RegisterSingleton[tracesdk.SpanExporter](c, observability.NewSpanExporter)
	ioc.RegisterSingleton[*observability.TracerProviderFactory](c, observability.NewTracerProviderFactory)
	ioc.RegisterSingleton[trace.TracerProvider](c, observability.NewDefaultTracerProvider)

	// register etcd
	ioc.RegisterSingleton[*etcd.Client](c, persistence.NewEtcdClient)

	// register shared state
	ioc.RegisterSingleton[persistence.State](c, persistence.NewEtcdState)

	// register services
	ioc.RegisterSingleton[service.ExhibitService](c, service.NewExhibitService)
	ioc.RegisterSingleton[service.ApplicationResolverService](c, service.NewDockerHostApplicationResolverService)

	// register livecheck
	ioc.RegisterSingleton[*impl.HttpLivecheck](c, func(resolverService service.ApplicationResolverService) *impl.HttpLivecheck {
		return &impl.HttpLivecheck{
			ApplicationResolverService: resolverService,
		}
	})
	ioc.RegisterSingleton[*impl.ExecLivecheck](c, util.IdentityF(new(impl.ExecLivecheck)))
	ioc.RegisterSingleton[service.LivecheckFactoryService](c, util.IdentityF(ioc.ForStruct[impl.LivecheckFactoryServiceImpl](c)))

	// register services
	ioc.RegisterSingleton[service.ApplicationProvisionerService](c, service.NewDockerApplicationProvisionerService)
	ioc.RegisterSingleton[service.ApplicationProvisionerHandlerService](c, service.NewApplicationProvisionerHandlerService)

	// register router and routes
	ioc.RegisterSingleton[*router.Mux](c, router.NewMux)
	ioc.ForFunc(c, health.RegisterRoutes)
	ioc.ForFunc(c, exhibit.RegisterRoutes)
	ioc.ForFunc(c, api.RegisterRoutes)

	//TODO: start cron goroutine to check for expired exhibits

	ioc.ForFunc(c, func(router *router.Mux, config config.Config, log *zap.SugaredLogger) {
		log.Infof("Starting server on port %s", config.GetPort())
		err := http.ListenAndServe(fmt.Sprintf(":%s", config.GetPort()), router)
		if err != nil {
			panic(err)
		}
	})

	graph := ioc.GenerateDependencyGraph(c)
	fmt.Println(graph)
}
