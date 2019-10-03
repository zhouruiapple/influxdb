package http

import (
	"net/http"

	"github.com/influxdata/influxdb/search"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"

	"github.com/influxdata/influxdb"
)

// SearchBackend is all services and associated parameters required to construct
// the SearchHandler.
type SearchBackend struct {
	Logger *zap.Logger
	influxdb.HTTPErrorHandler

	FindService search.FindService
}

// NewSearchBackend returns a new instance of SearchBackend.
func NewSearchBackend(b *APIBackend) *SearchBackend {
	return &SearchBackend{
		HTTPErrorHandler: b.HTTPErrorHandler,
		Logger:           b.Logger.With(zap.String("handler", "bucket")),

		FindService: b.FindService,
	}
}

// SearchHandler represents an HTTP API handler for buckets.
type SearchHandler struct {
	*httprouter.Router
	influxdb.HTTPErrorHandler
	Logger *zap.Logger

	FindService search.FindService
}

const (
	searchPath = "/api/v2/search"
)

// NewSearchHandler returns a new instance of SearchHandler.
func NewSearchHandler(b *SearchBackend) *SearchHandler {
	h := &SearchHandler{
		Router:           NewRouter(b.HTTPErrorHandler),
		HTTPErrorHandler: b.HTTPErrorHandler,
		Logger:           b.Logger,

		FindService: b.FindService,
	}

	h.HandlerFunc("GET", searchPath, h.handleGetSearch)

	return h
}

func (h *SearchHandler) handleGetSearch(w http.ResponseWriter, r *http.Request) {
	// pull auth from ctx, populate OwnerID
	ctx := r.Context()
	docType := r.URL.Query().Get("docType")
	q := r.URL.Query().Get("q")
	docs, err := h.FindService.SimpleQuery(q, search.DocType(docType))
	if err != nil {
		h.HandleHTTPError(ctx, err, w)
		return
	}
	if err := encodeResponse(ctx, w, http.StatusOK, docs); err != nil {
		logEncodingError(h.Logger, r, err)
		return
	}

}
