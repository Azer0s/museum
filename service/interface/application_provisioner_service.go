package service

type ApplicationProvisionerService interface {
	StartApplication(exhibitId string) error
	StopApplication(exhibitId string) error
	CleanupApplication(exhibitId string) error
}
