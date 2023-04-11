package service

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go/v2/event"
)

type ApplicationProvisionerHandlerService interface {
	HandleEvent(ctx context.Context, event *cloudevents.Event, exhibitId string) error
}
