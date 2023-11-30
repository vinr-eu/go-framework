package code

type Code struct {
	code string
	text string
}

func NewCode(code string, text string) Code {
	return Code{code: code, text: text}
}

func (e Code) GetCode() string {
	return e.code
}

func (e Code) GetText() string {
	return e.text
}

func (e Code) String() string {
	return e.code
}
