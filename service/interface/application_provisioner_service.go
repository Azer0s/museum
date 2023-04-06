package service

import "context"

type ApplicationProvisionerService interface {
	StartApplication(ctx context.Context, exhibitId string) error
	StopApplication(ctx context.Context, exhibitId string) error
	CleanupApplication(ctx context.Context, exhibitId string) error
}
