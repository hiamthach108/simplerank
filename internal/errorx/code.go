package errorx

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
