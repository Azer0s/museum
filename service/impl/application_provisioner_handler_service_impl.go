package impl

import (
	"context"
	"errors"
	cloudevents "github.com/cloudevents/sdk-go/v2/event"
	"museum/domain"
	service "museum/service/interface"
)

type ApplicationProvisionerHandlerServiceImpl struct {
	ApplicationProvisionerService service.ApplicationProvisionerService
}

func (a ApplicationProvisionerHandlerServiceImpl) HandleEvent(_ context.Context, event *cloudevents.Event, _ string) error {
	switch event.Type() {
	case domain.StartEventType:
		// TODO
		break
	case domain.StopEventType:
		// TODO
		break
	default:
		return errors.New("unknown event type " + event.Type())
	}

	return nil
}
