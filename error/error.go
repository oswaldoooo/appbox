package error

type Error interface {
	Code() uint8
	error
}

type Err struct{}

func (e *Err) Code() uint8 {
	panic("not implemented") // TODO: Implement
}

func (e *Err) Error() string {
	panic("not implemented") // TODO: Implement
}
