package rotatelogs

type Interface interface {
	Name() string
	Value() any
}

type Option struct {
	name  string
	value any
}

func NewOption(name string, value any) *Option {
	return &Option{
		name:  name,
		value: value,
	}
}

func (o *Option) Name() string {
	return o.name
}

func (o *Option) Value() any {
	return o.value
}
