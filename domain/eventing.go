package domain

type ObjectStartingStep int

func (o ObjectStartingStep) String() string {
	return [...]string{"clean", "create", "start", "livecheck", "ready"}[o]
}

const (
	ObjectStartingStepClean ObjectStartingStep = iota
	ObjectStartingStepCreate
	ObjectStartingStepStart
	ObjectStartingStepLivecheck
	ObjectStartingStepReady
)

type ExhibitStartingStep struct {
	//index of the object in the exhibit
	Object     int
	TotalSteps int
	Step       ObjectStartingStep
}

type ExhibitMetadata struct {
	ExhibitId          string
	ExhibitName        string
	ExhibitDescription string
	Meta               map[string]string
}
