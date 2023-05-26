package server

import (
	"context"
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
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	ctx := context.Background()
	signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGSTOP)

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
	ioc.RegisterSingleton[service.LockService](c, service.NewLockService)
	ioc.RegisterSingleton[service.RuntimeInfoService](c, service.NewRuntimeInfoService)
	ioc.RegisterSingleton[service.ExhibitService](c, service.NewExhibitService)
	ioc.RegisterSingleton[service.LastAccessedService](c, service.NewLastAccessedService)
	ioc.RegisterSingleton[service.ApplicationResolverService](c, service.NewDockerHostApplicationResolverService)

	// register livecheck
	ioc.RegisterSingleton[*service.HttpLivecheck](c, service.NewHttpLivecheck)
	ioc.RegisterSingleton[*service.ExecLivecheck](c, service.NewExecLivecheck)
	ioc.RegisterSingleton[service.LivecheckFactoryService](c, service.NewLivecheckFactoryService)

	// register services
	ioc.RegisterSingleton[service.ApplicationProvisionerService](c, service.NewDockerApplicationProvisionerService)
	ioc.RegisterSingleton[service.ApplicationProvisionerHandlerService](c, service.NewApplicationProvisionerHandlerService)
	ioc.RegisterSingleton[service.ExhibitCleanupService](c, service.NewExhibitCleanupService)

	// register router and routes
	ioc.RegisterSingleton[*router.Mux](c, router.NewMux)
	ioc.ForFunc(c, health.RegisterRoutes)
	ioc.ForFunc(c, exhibit.RegisterRoutes)
	ioc.ForFunc(c, api.RegisterRoutes)

	go ioc.ForFunc(c, func(router *router.Mux, config config.Config, log *zap.SugaredLogger) {
		log.Infof("starting server on port %s", config.GetPort())
		err := http.ListenAndServe(fmt.Sprintf(":%s", config.GetPort()), router)
		if err != nil {
			panic(err)
		}
	})

	go ioc.ForFunc(c, func(log *zap.SugaredLogger, cleanupService service.ExhibitCleanupService) {
		for {
			<-time.After(10 * time.Second)

			log.Info("checking for expired exhibits")

			err := cleanupService.Cleanup()
			if err != nil {
				log.Errorw("failed to cleanup exhibits", "error", err)
			}
		}
	})

	<-ctx.Done()
}
