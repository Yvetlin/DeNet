package middleware

import (
	"log"
	"net/http"

	"DeNet/utils"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				var errObj error
				if e, ok := err.(error); ok {
					errObj = e
				} else {
					errObj = &panicError{message: "Internal server error"}
				}
				utils.RespondWithInternalError(w, errObj)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type panicError struct {
	message string
}

func (e *panicError) Error() string {
	return e.message
}

