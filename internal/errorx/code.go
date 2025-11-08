package errorx

import "fmt"

type AppError struct {
	Code    AppErrCode
	Message string
	Err     error // optional underlying error
}

type AppErrCode int

const (
	// General errors
	ErrInternal      AppErrCode = 500
	ErrBadRequest    AppErrCode = 400
	ErrNotFound      AppErrCode = 404
	ErrUnauthorized  AppErrCode = 401
	ErrForbidden     AppErrCode = 403
	ErrConflict      AppErrCode = 409
	ErrUnprocessable AppErrCode = 422
	ErrRateLimit     AppErrCode = 429

	// Leaderboard errors
	ErrLeaderboardNotFound AppErrCode = 1001
	ErrInvalidEntry        AppErrCode = 1002
	ErrCreateLeaderboard   AppErrCode = 1003
	ErrUpdateLeaderboard   AppErrCode = 1004
	ErrUpdateScore         AppErrCode = 1005
)

var errorMsgs = map[AppErrCode]string{
	ErrInternal:      "Internal server error",
	ErrBadRequest:    "Bad request",
	ErrNotFound:      "Resource not found",
	ErrUnauthorized:  "Unauthorized access",
	ErrForbidden:     "Forbidden access",
	ErrConflict:      "Resource conflict",
	ErrUnprocessable: "Unprocessable entity",
	ErrRateLimit:     "Too many requests",

	ErrLeaderboardNotFound: "Leaderboard not found",
	ErrInvalidEntry:        "Invalid leaderboard entry",
	ErrCreateLeaderboard:   "Failed to create leaderboard",
	ErrUpdateLeaderboard:   "Failed to update leaderboard",
	ErrUpdateScore:         "Failed to update score",
}

// GetErrorMessage returns a user-friendly error message for a given error code.
func GetErrorMessage(code int) string {
	if msg, exists := errorMsgs[AppErrCode(code)]; exists {
		return msg
	}
	return "An unknown error occurred."
}

// Implement the `error` interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Wrap creates a new AppError with an existing error
func Wrap(code AppErrCode, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: GetErrorMessage(int(code)),
		Err:     err,
	}
}

// New creates a new AppError with custom message
func New(code AppErrCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}
