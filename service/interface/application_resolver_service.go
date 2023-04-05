package service

type ApplicationResolverService interface {
	ResolveApplication(exhibitId string) (string, error)
}
