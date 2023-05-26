package impl

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/domain"
	service "museum/service/interface"
	"time"
)

type ExhibitCleanupServiceImpl struct {
	ExhibitService                service.ExhibitService
	LockService                   service.LockService
	ApplicationProvisionerService service.ApplicationProvisionerService
	Provider                      trace.TracerProvider
	Log                           *zap.SugaredLogger
}

func (e ExhibitCleanupServiceImpl) cleanupExhibit(exhibit domain.Exhibit, ctx context.Context) {
	subCtx, span := e.Provider.
		Tracer("cleanup-service").
		Start(ctx, "CleanupExhibit("+exhibit.Id+")", trace.WithAttributes(attribute.String("exhibitId", exhibit.Id)))
	defer span.End()

	e.Log.Debugw("checking exhibit", "exhibitId", exhibit.Id)

	duration, err := time.ParseDuration(exhibit.Lease)
	if err != nil {
		e.Log.Warnw("error parsing lease duration", "error", err, "exhibitId", exhibit.Id)
		return
	}

	if time.Now().After(time.Unix(exhibit.RuntimeInfo.LastAccessed, 0).Add(duration)) && exhibit.RuntimeInfo.Status == domain.Running {
		expiredBy := time.Now().Sub(time.Unix(exhibit.RuntimeInfo.LastAccessed, 0).Add(duration)).String()
		e.Log.Infow("exhibit lease expired", "exhibitId", exhibit.Id, "expiredBy", expiredBy)
		span.AddEvent("exhibit lease expired by " + expiredBy + ", cleaning up")

		err = e.ApplicationProvisionerService.StopApplication(subCtx, exhibit.Id)
		if err != nil {
			e.Log.Warnw("error stopping application", "error", err, "exhibitId", exhibit.Id)
			return
		}

		err = e.ApplicationProvisionerService.CleanupApplication(subCtx, exhibit.Id)
		if err != nil {
			e.Log.Warnw("error cleaning up application", "error", err, "exhibitId", exhibit.Id)
			return
		}

		e.Log.Infow("exhibit cleaned up", "exhibitId", exhibit.Id)
	}
}

func (e ExhibitCleanupServiceImpl) Cleanup() error {
	ctx, span := e.Provider.
		Tracer("cleanup-service").
		Start(context.Background(), "Cleanup")
	defer span.End()

	span.AddEvent("getting all exhibits")

	exhibits := e.ExhibitService.GetAllExhibits(ctx)
	for _, exhibit := range exhibits {
		span.AddEvent("checking exhibit " + exhibit.Id)
		e.cleanupExhibit(exhibit, ctx)
	}

	//TODO: cleanup starting exhibits that have been starting for too long

	return nil
}
