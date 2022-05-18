package timers

import (
	"net/http"

	"github.com/felixge/httpsnoop"
)

type MiddlewareOptions struct {
	Callback func(*TimerSet) // This function will be called at the end of the request
}

func Middleware(next http.Handler, opts MiddlewareOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r.Context())
		r = r.WithContext(ctx)

		// TimerSet.AddHeader() always adds a Server-Timing header, because users might want
		// to write out multiple timings, one for each set. So we need to make sure we don't
		// duplicate things ourselves.
		headerAdded := false

		// Hook into both WriteHeader and Write, to add our header just before it's too late.
		h := httpsnoop.Hooks{
			WriteHeader: func(original httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
				return func(code int) {
					headerAdded = true
					From(ctx).AddHeader(w)
					original(code)
				}
			},
			Write: func(original httpsnoop.WriteFunc) httpsnoop.WriteFunc {
				return func(b []byte) (int, error) {
					headerAdded = true
					From(ctx).AddHeader(w)
					return original(b)
				}
			},
		}
		w = httpsnoop.Wrap(w, h)
		next.ServeHTTP(w, r)
		if !headerAdded {
			From(ctx).AddHeader(w)
		}
		if opts.Callback != nil {
			opts.Callback(From(ctx))
		}
	})
}
