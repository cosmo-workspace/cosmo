package kosmo

var (
	ErrIsAlreadyExists     = &KosmoError{}
	ErrNotFound            = &KosmoError{}
	ErrBadRequest          = &KosmoError{}
	ErrForbidden           = &KosmoError{}
	ErrUnauthorized        = &KosmoError{}
	ErrInternalServerError = &KosmoError{}
	ErrServiceUnavailable  = &KosmoError{}
)

type KosmoError struct {
	message   string
	err       error
	errorType error
}

func newKosmoError(errorType *KosmoError, message string, err error) *KosmoError {
	return &KosmoError{
		message:   message,
		err:       err,
		errorType: errorType,
	}
}

func (e *KosmoError) Error() string {
	return e.message
}

func (e *KosmoError) Is(target error) bool {
	return e.errorType == target
}

func (e *KosmoError) Unwrap() error {
	return e.err
}

func NewNotFoundError(message string, err error) error {
	return newKosmoError(ErrNotFound, message, err)
}

func NewIsAlreadyExistsError(message string, err error) error {
	return newKosmoError(ErrIsAlreadyExists, message, err)
}

func NewBadRequestError(message string, err error) error {
	return newKosmoError(ErrBadRequest, message, err)
}

func NewForbiddenError(message string, err error) error {
	return newKosmoError(ErrForbidden, message, err)
}

func NewUnauthorizedError(message string, err error) error {
	return newKosmoError(ErrUnauthorized, message, err)
}

func NewInternalServerError(message string, err error) error {
	return newKosmoError(ErrInternalServerError, message, err)
}

func NewServiceUnavailableError(message string, err error) error {
	return newKosmoError(ErrServiceUnavailable, message, err)
}
