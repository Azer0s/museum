package service

type ApplicationProvisionService interface {
	StartApplication(exhibitId string) error
	StopApplication(exhibitId string) error
	CleanupApplication(exhibitId string) error
}
