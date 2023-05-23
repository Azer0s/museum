package impl

import service "museum/service/interface"

type LivecheckFactoryServiceImpl struct {
	HttpLivecheck *HttpLivecheck
	ExecLivecheck *ExecLivecheck
}

func (l *LivecheckFactoryServiceImpl) GetLivecheckService(objectType string) service.LivecheckService {
	switch objectType {
	case "http":
		return l.HttpLivecheck
	case "exec":
		return l.ExecLivecheck
	default:
		return nil
	}
}
