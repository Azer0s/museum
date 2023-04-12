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

func (a ApplicationProvisionerHandlerServiceImpl) HandleEvent(ctx context.Context, event *cloudevents.Event, id string) error {
	switch event.Type() {
	case domain.StartEventType:
		err := a.ApplicationProvisionerService.StartApplication(ctx, id)
		if err != nil {
			return err
		}
	case domain.StopEventType:
		err := a.ApplicationProvisionerService.StopApplication(ctx, id)
		if err != nil {
			return err
		}
	default:
		return errors.New("unknown event type " + event.Type())
	}

	return nil
}
