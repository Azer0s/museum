package impl

import (
	service "museum/service/interface"
)

type LivecheckFactoryServiceImpl struct {
	HttpLivecheck service.Livecheck
	ExecLivecheck service.Livecheck
}

func (l *LivecheckFactoryServiceImpl) GetLivecheckService(objectType string) service.Livecheck {
	switch objectType {
	case "http":
		return l.HttpLivecheck
	case "exec":
		return l.ExecLivecheck
	default:
		return nil
	}
}
