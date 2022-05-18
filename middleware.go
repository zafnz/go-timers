package timers

import (
	"net/http"

	"github.com/felixge/httpsnoop"
)

type MiddlewareOptions struct {
	Callback       func(*TimerSet) // This function will be called at the end of the request
	NoDefaultTimer bool            // If true, then no default timer will be set.
}

// The middleware function sets up timers for each request, and for each request emits
// a Server-Timing header. A default "Request" timer is created, unless the option
// NoDefaultTimer is true. Use this function as you would any other middleware function
//  handler = timers.Middleware(handler, MiddlewareOptions{})
func Middleware(next http.Handler, opts MiddlewareOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r.Context())
		r = r.WithContext(ctx)
		var t *Timer

		if opts.NoDefaultTimer {
			t = &Timer{}
		} else {
			t = From(ctx).New("Request").Start()
		}

		// TimerSet.AddHeader() always adds a Server-Timing header, because users might want
		// to write out multiple timings, one for each set. So we need to make sure we don't
		// duplicate things ourselves.
		headerAdded := false

		// Hook into both WriteHeader and Write, to add our header just before it's too late.
		h := httpsnoop.Hooks{
			WriteHeader: func(original httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
				return func(code int) {
					headerAdded = true
					t.Stop()
					From(ctx).AddHeader(w)
					original(code)
				}
			},
			Write: func(original httpsnoop.WriteFunc) httpsnoop.WriteFunc {
				return func(b []byte) (int, error) {
					headerAdded = true
					t.Stop()
					From(ctx).AddHeader(w)
					return original(b)
				}
			},
		}
		w = httpsnoop.Wrap(w, h)
		next.ServeHTTP(w, r)
		if !headerAdded {
			t.Stop()
			From(ctx).AddHeader(w)
		}
		if opts.Callback != nil {
			opts.Callback(From(ctx))
		}
	})
}
