package service

import (
	docker "github.com/docker/docker/client"
	"museum/service/impl"
	service "museum/service/interface"
)

type LivecheckFactoryService service.LivecheckFactoryService
type LivecheckFactoryServiceImpl impl.LivecheckFactoryServiceImpl
type HttpLivecheck impl.HttpLivecheck
type ExecLivecheck impl.ExecLivecheck

func NewHttpLivecheck(applicationResolverService service.ApplicationResolverService, exhibitService service.ExhibitService) *HttpLivecheck {
	return (*HttpLivecheck)(&impl.HttpLivecheck{
		ApplicationResolverService: applicationResolverService,
		ExhibitService:             exhibitService,
	})
}

func NewExecLivecheck(client *docker.Client) *ExecLivecheck {
	return (*ExecLivecheck)(&impl.ExecLivecheck{
		Client: client,
	})
}

func NewLivecheckFactoryService(httpLivecheck *HttpLivecheck, execLivecheck *ExecLivecheck) LivecheckFactoryService {
	return &impl.LivecheckFactoryServiceImpl{
		HttpLivecheck: (*impl.HttpLivecheck)(httpLivecheck),
		ExecLivecheck: (*impl.ExecLivecheck)(execLivecheck),
	}
}
