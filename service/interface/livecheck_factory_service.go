package service

type LivecheckFactoryService interface {
	GetLivecheckService(objectType string) LivecheckService
}
