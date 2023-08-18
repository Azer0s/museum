package domain

type ObjectStartingStep int

const (
	ObjectStartingStepClean ObjectStartingStep = iota
	ObjectStartingStepCreate
	ObjectStartingStepStart
	ObjectStartingStepLivecheck
	ObjectStartingStepReady
)

type ExhibitStartingStep struct {
	//index of the object in the exhibit
	Object int
	Step   ObjectStartingStep
}

type ExhibitMetadata struct {
	ExhibitId          string
	ExhibitName        string
	ExhibitDescription string
	Meta               map[string]string
}
