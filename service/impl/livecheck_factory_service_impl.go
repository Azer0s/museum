package impl

import (
	"museum/domain"
	service "museum/service/interface"
)

type LivecheckFactoryServiceImpl struct {
	HttpLivecheck service.Livecheck
	ExecLivecheck service.Livecheck
}

func (l *LivecheckFactoryServiceImpl) GetLivecheckService(objectType string) service.Livecheck {
	switch objectType {
	case domain.LivecheckTypeHttp:
		return l.HttpLivecheck
	case domain.LivecheckTypeExec:
		return l.ExecLivecheck
	default:
		return nil
	}
}
