package spinner

type NullSpinner struct {
	currentHeading string
}

func NewNullSpinner() *NullSpinner {
	return &NullSpinner{}
}

func (s *NullSpinner) Active() bool {
	return false
}

func (s *NullSpinner) Start() {}

func (s *NullSpinner) Stop() {}

func (s *NullSpinner) StampShadow() {}

func (s *NullSpinner) Success(message string) {
}

func (s *NullSpinner) Error(message string) {
}

func (s *NullSpinner) Skip(message string) {
}

func (s *NullSpinner) SetMessage(message string) {
}

func (s *NullSpinner) NewStep(message string) {
}

func (s *NullSpinner) HideAndExecute(f func()) {

}

func (s *NullSpinner) CurrentHeading() string {
	return s.currentHeading
}
