package domain

import "strconv"

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
	Object int
	Step   ObjectStartingStep
	Error  error
}

type ExhibitStartingStepEvent struct {
	ExhibitId        string `json:"exhibitId"`
	Object           string `json:"object"`
	Step             string `json:"step"`
	Error            string `json:"error"`
	CurrentStepCount int    `json:"currentStepCount"`
	TotalStepCount   int    `json:"totalStepCount"`
}

type ExhibitStoppingEvent struct {
	ExhibitId string `json:"exhibitId"`
}

func (e ExhibitStartingStepEvent) ToMap() map[string]string {
	return map[string]string{
		"exhibitId":        e.ExhibitId,
		"object":           e.Object,
		"step":             e.Step,
		"currentStepCount": strconv.Itoa(e.CurrentStepCount),
		"totalStepCount":   strconv.Itoa(e.TotalStepCount),
	}
}

type ExhibitMetadata struct {
	ExhibitId          string
	ExhibitName        string
	ExhibitDescription string
	Meta               map[string]string
}
