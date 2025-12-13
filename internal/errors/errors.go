package errors

// Error codes for the order service
const (
	CodeUnknown           string = "ORD_UNKNOWN"
	CodeOrderNotFound     string = "ORD_ORDER_NOT_FOUND"
	CodeOrderCreateFailed string = "ORD_CREATE_FAILED"
	CodeOrderUpdateFailed string = "ORD_UPDATE_FAILED"
	CodeInvalidInput      string = "ORD_INVALID_INPUT"
	CodeInvalidStatus     string = "ORD_INVALID_STATUS"
	CodeInvalidProduct    string = "ORD_INVALID_PRODUCT"
	CodeInsufficientStock string = "ORD_INSUFFICIENT_STOCK"
	CodePaymentFailed     string = "ORD_PAYMENT_FAILED"
	CodeUnauthorized      string = "ORD_UNAUTHORIZED"
	CodeInternalError     string = "ORD_INTERNAL_ERROR"
	CodeDatabaseError     string = "ORD_DATABASE_ERROR"
	CodeKafkaError        string = "ORD_KAFKA_ERROR"
)

// OrderError represents a custom error with an error code
type OrderError struct {
	ErrorCode string
	Message   string
}

func (e *OrderError) Error() string {
	return e.Message
}

// Predefined errors
var (
	ErrOrderNotFound     = &OrderError{ErrorCode: CodeOrderNotFound, Message: "order not found"}
	ErrOrderCreateFailed = &OrderError{ErrorCode: CodeOrderCreateFailed, Message: "failed to create order"}
	ErrOrderUpdateFailed = &OrderError{ErrorCode: CodeOrderUpdateFailed, Message: "failed to update order"}
	ErrInvalidInput      = &OrderError{ErrorCode: CodeInvalidInput, Message: "invalid input"}
	ErrInvalidStatus     = &OrderError{ErrorCode: CodeInvalidStatus, Message: "invalid order status"}
	ErrInvalidProduct    = &OrderError{ErrorCode: CodeInvalidProduct, Message: "invalid product"}
	ErrInsufficientStock = &OrderError{ErrorCode: CodeInsufficientStock, Message: "insufficient stock"}
	ErrPaymentFailed     = &OrderError{ErrorCode: CodePaymentFailed, Message: "payment failed"}
	ErrUnauthorized      = &OrderError{ErrorCode: CodeUnauthorized, Message: "unauthorized"}
	ErrInternalError     = &OrderError{ErrorCode: CodeInternalError, Message: "internal server error"}
	ErrDatabaseError     = &OrderError{ErrorCode: CodeDatabaseError, Message: "database error"}
	ErrKafkaError        = &OrderError{ErrorCode: CodeKafkaError, Message: "kafka error"}
)

// NewOrderError creates a new OrderError with a custom message
func NewOrderError(code string, message string) *OrderError {
	return &OrderError{
		ErrorCode: code,
		Message:   message,
	}
}

// IsOrderError checks if an error is an OrderError
func IsOrderError(err error) bool {
	_, ok := err.(*OrderError)
	return ok
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) string {
	if orderErr, ok := err.(*OrderError); ok {
		return orderErr.ErrorCode
	}
	return CodeUnknown
}

const DefaultErrorMessage = "an unexpected error occurred"

// ParseError extracts the error code and message from any error.
// If the error is not an OrderError, returns CodeUnknown and default message.
func ParseError(err error) (string, string) {
	if err == nil {
		return CodeUnknown, DefaultErrorMessage
	}

	if orderErr, ok := err.(*OrderError); ok {
		return orderErr.ErrorCode, orderErr.Message
	}

	return CodeUnknown, DefaultErrorMessage
}

// GetError converts any error to OrderError
func GetError(err error) *OrderError {
	if err == nil {
		return nil
	}

	if orderErr, ok := err.(*OrderError); ok {
		return orderErr
	}

	return &OrderError{
		ErrorCode: CodeUnknown,
		Message:   DefaultErrorMessage,
	}
}

// Wrap wraps an error with a custom OrderError
func Wrap(code string, err error) *OrderError {
	if err == nil {
		return nil
	}
	return &OrderError{
		ErrorCode: code,
		Message:   err.Error(),
	}
}
