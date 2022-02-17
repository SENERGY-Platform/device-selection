package basecontentvariable

type Descriptor interface {
	GetName() string
	GetCharacteristicId() string
	GetSubContentVariables() []Descriptor
	GetFunctionId() string
	GetAspectId() string
}
