package v1

const (
	NotReady             = "NotReady"
	Running              = "Running"
	Creating             = "Creating"
	PodNotFound          = "PodNotFound"
	ContainerStatusIsNil = "ContainerStatusIsNil"
	PodNotRunning        = "PodNotRunning"
)

type RegistryErrors struct {
	errorType    *string
	errorMessage *string
}

func MakeRegistryError(e string) error {
	RegistryError := RegistryErrors{}
	if e == NotReady || e == Running || e == Creating || e == PodNotFound || e == ContainerStatusIsNil || e == PodNotRunning {
		RegistryError.errorType = &e
	} else {
		RegistryError.errorMessage = &e
	}
	return RegistryError
}

func (r RegistryErrors) Error() string {
	if r.errorType != nil {
		return *r.errorType
	}

	return *r.errorMessage
}

func IsPodError(err error) bool {
	if err.Error() == PodNotFound || err.Error() == ContainerStatusIsNil || err.Error() == PodNotRunning {
		return true
	}

	return false
}
