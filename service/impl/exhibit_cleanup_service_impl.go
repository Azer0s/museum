package impl

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/domain"
	service "museum/service/interface"
	"museum/util"
	"time"
)

type ExhibitCleanupServiceImpl struct {
	ExhibitService service.ExhibitService
	LockService    service.LockService
	Provider       trace.TracerProvider
	Log            *zap.SugaredLogger
}

func (e ExhibitCleanupServiceImpl) cleanupExhibit(exhibit domain.Exhibit, ctx context.Context) {
	subCtx, span := e.Provider.
		Tracer("cleanup-service").
		Start(ctx, "CleanupExhibit("+exhibit.Id+")", trace.WithAttributes(attribute.String("exhibitId", exhibit.Id)))
	defer span.End()

	span.AddEvent("acquiring exhibit lock")

	lock := e.LockService.GetRwLock(subCtx, exhibit.Id, "exhibit")
	err := lock.Lock()
	if err != nil {
		e.Log.Errorw("error locking exhibit lock", "error", err, "exhibitId", exhibit.Id)
		return
	}

	span.AddEvent("exhibit lock acquired")

	defer func(lock util.RwErrMutex) {
		err := lock.Unlock()
		if err != nil {
			e.Log.Errorw("error unlocking exhibit lock", "error", err, "exhibitId", exhibit.Id)
		}
	}(lock)

	e.Log.Debugw("checking exhibit", "exhibitId", exhibit.Id)

	duration, err := time.ParseDuration(exhibit.Lease)
	if err != nil {
		e.Log.Warnw("error parsing lease duration", "error", err, "exhibitId", exhibit.Id)
		return
	}

	if time.Now().After(time.Unix(exhibit.RuntimeInfo.LastAccessed, 0).Add(duration)) {
		expiredBy := time.Now().Sub(time.Unix(exhibit.RuntimeInfo.LastAccessed, 0).Add(duration)).String()
		e.Log.Infow("exhibit lease expired", "exhibitId", exhibit.Id, "expiredBy", expiredBy)
		span.AddEvent("exhibit lease expired by " + expiredBy + ", cleaning up")
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

	return nil
}
