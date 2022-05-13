package timers

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed waterfall waterfall/index.html
var content embed.FS

// Returns a http.Handler that will serve the waterfall inspector suitable for
// rendering a waterfall of the Server-Timing header.
// When registering the handler, ensure you strip prefix. Eg:
//  http.Handle("/waterfall/", http.StripPrefix("/waterfall/", timers.WaterfallHandler()))
func WaterfallHandler() http.Handler {
	fsys, err := fs.Sub(content, "waterfall")
	if err != nil {
		log.Fatal(err)
	}
	return http.FileServer(http.FS(fsys))
}
