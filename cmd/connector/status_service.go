package main

const (
	Unset      int = 0
	Irrelevant int = 4
	Running    int = 1
	Fail       int = 2
	Successful int = 3

	UnsetString      string = "UNSET"
	IrrelevantString string = "NOT_RELEVANT"
	RunningString    string = "SCHEDULED"
	FailString       string = "FAILED"
	SuccessfulString string = "SCHEDULED"
)

var (
	statusUnset      = StatusServiceImpl{Unset}
	statusIrrelevant = StatusServiceImpl{Irrelevant}
	statusRunning    = StatusServiceImpl{Running}
	statusFail       = StatusServiceImpl{Fail}
	statusSuccessful = StatusServiceImpl{Successful}
)

type StatusService interface {
	String() string
}

// status encodes the checker states.
type StatusServiceImpl struct {
	Status int
}

func (s StatusServiceImpl) String() string {
	return map[StatusServiceImpl]string{
		statusUnset:      UnsetString,
		statusIrrelevant: IrrelevantString,
		statusRunning:    RunningString,
		statusFail:       FailString,
		// remember - success here, simply means we have successfully informed the event listener of the job...
		statusSuccessful: SuccessfulString,
	}[s]
}
