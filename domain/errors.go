package domain

type BusinessErr struct {
	Msg string
}

func (e BusinessErr) Error() string {
	return e.Msg
}

type NameAlreadyInUseErr struct {
	Msg string
}

func (e NameAlreadyInUseErr) Error() string {
	return e.Msg
}

type AddressNotValidErr struct {
	Msg string
}

func (e AddressNotValidErr) Error() string {
	return e.Msg
}
