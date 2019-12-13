package http

import (
	"net/http"
	"strconv"

	"github.com/influxdata/flux/iocounter"
	"github.com/influxdata/httprouter"
	platform "github.com/influxdata/influxdb"
	idpctx "github.com/influxdata/influxdb/context"
	"github.com/influxdata/influxdb/query"
	"github.com/influxdata/influxdb/query/influxql"
	"go.uber.org/zap"
)

// InfluxqlHandler mimics the /query handler from influxdb, but, enriches
// with org and forwards requests to the transpiler service.
type InfluxqlHandler struct {
	*httprouter.Router

	*InfluxQLBackend
}

type InfluxQLBackend struct {
	platform.HTTPErrorHandler
	Logger               *zap.Logger
	AuthorizationService platform.AuthorizationService
	OrganizationService  platform.OrganizationService
	ProxyQueryService    query.ProxyQueryService
}

// NewInfluxQLBackend constructs an InfluxQLBackend from a LegacyBackend.
func NewInfluxQLBackend(b *APIBackend) *InfluxQLBackend {
	return &InfluxQLBackend{
		HTTPErrorHandler:     b.HTTPErrorHandler,
		Logger:               b.Logger.With(zap.String("handler", "influxql")),
		AuthorizationService: b.AuthorizationService,
		OrganizationService:  b.OrganizationService,
		ProxyQueryService:    b.InfluxQLService,
	}
}

// NewInfluxQLHandler returns a new instance of InfluxqlHandler to handle
// influxql v1 queries
func NewInfluxQLHandler(b *InfluxQLBackend) *InfluxqlHandler {
	h := &InfluxqlHandler{
		Router:          httprouter.New(),
		InfluxQLBackend: b,
	}

	h.HandlerFunc("GET", "/query", h.HandleQuery)
	h.HandlerFunc("POST", "/query", h.HandleQuery)
	return h
}

// HandleQuery mimics the influxdb 1.0 /query
func (h *InfluxqlHandler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	panic("hahah")
	ctx := r.Context()
	defer r.Body.Close()

	a, err := idpctx.GetAuthorizer(ctx)
	if err != nil {
		h.HandleHTTPError(ctx, err, w)
		return
	}

	active := false
	switch a.Kind() {
	case "authorization":
		active = a.(*platform.Authorization).IsActive()
	case "session":
		h.HandleHTTPError(ctx, &platform.Error{
			Code: platform.EForbidden,
			Msg:  "insufficient permissions; session not supported",
		}, w)
		return
	}

	if !active {
		h.HandleHTTPError(ctx, &platform.Error{
			Code: platform.EForbidden,
			Msg:  "insufficient permissions",
		}, w)
		return
	}

	analyze(r, h.Logger)
	compiler := decodeInfluxCompiler(r)
	dialect := decodeInfluxDialect(r)

	o, err := h.OrganizationService.FindOrganization(ctx, platform.OrganizationFilter{
		Name: &compiler.Cluster,
	})
	if err != nil {
		h.HandleHTTPError(ctx, err, w)
		return
	}

	switch dialect.Encoding {
	case influxql.JSON, influxql.JSONPretty:
		w.Header().Set("Content-Type", "application/json")
	case influxql.CSV:
		w.Header().Set("Content-Type", "text/csv")
	case influxql.Msgpack:
		w.Header().Add("Content-Type", "application/x-msgpack")
	}

	req := &query.ProxyRequest{
		Request: query.Request{
			Authorization:  a.(*platform.Authorization),
			OrganizationID: o.ID,
			Compiler:       compiler,
		},
		Dialect: dialect,
	}

	cw := iocounter.Writer{Writer: w}
	_, err = h.ProxyQueryService.Query(ctx, &cw, req)
	n := cw.Count()

	if err != nil {
		if n == 0 {
			// Only record the error headers IFF nothing has been written to w.
			h.HandleHTTPError(ctx, err, w)
			return
		}
		h.Logger.Info("error writing response to client",
			zap.String("org", o.Name),
			zap.String("handler", "influxql"),
			zap.Error(err),
		)
	}
}

// DefaultChunkSize is the default number of points to write in
// one chunk.
const DefaultChunkSize = 10000

func decodeInfluxDialect(r *http.Request) *influxql.Dialect {
	qp := r.URL.Query()
	dialect := &influxql.Dialect{}

	if qp.Get("chunked") == "true" {
		dialect.ChunkSize = DefaultChunkSize
		size, err := strconv.Atoi(qp.Get("chunk_size"))
		if err == nil && size > 0 {
			dialect.ChunkSize = size
		}
	}

	switch qp.Get("epoch") {
	case "":
		dialect.TimeFormat = influxql.RFC3339Nano
	case "h":
		dialect.TimeFormat = influxql.Hour
	case "m":
		dialect.TimeFormat = influxql.Minute
	case "s":
		dialect.TimeFormat = influxql.Second
	case "ms":
		dialect.TimeFormat = influxql.Millisecond
	case "u", "us":
		dialect.TimeFormat = influxql.Microsecond
	default: // InfluxDB would default to nanoseconds if there was any string.
		dialect.TimeFormat = influxql.Nanosecond
	}

	switch r.Header.Get("Accept") {
	case "application/csv", "text/csv":
		dialect.Encoding = influxql.CSV
	case "application/x-msgpack":
		dialect.Encoding = influxql.Msgpack
	default:
		dialect.Encoding = influxql.JSON
		if qp.Get("pretty") == "true" {
			dialect.Encoding = influxql.JSONPretty
		}
	}

	if r.Header.Get("Content-Encoding") == "gzip" {
		dialect.Compression = influxql.Gzip
	}

	return dialect
}

func decodeInfluxCompiler(r *http.Request) *influxql.Compiler {
	compiler := &influxql.Compiler{
		Cluster: r.FormValue("org"),
		DB:      r.FormValue("db"),
		RP:      r.FormValue("rp"),
		Query:   r.FormValue("q"),
	}

	return compiler
}

// analyze is just to watch differences between this implementation and influxdb 1
func analyze(r *http.Request, logger *zap.Logger) {
	for param := range r.URL.Query() {
		switch param {
		case "org", "q", "db", "rp", "epoch", "chunk_size", "pretty", "chunked":
		default:
			logger.Debug("unhandled parameter", zap.String("handler", "influxql"), zap.String("param", param))
		}
	}
	switch r.Header.Get("Accept") {
	case "", "application/json":
	default:
		logger.Debug("unhandled content format", zap.String("handler", "influxql"), zap.String("format", r.Header.Get("Accept")))
	}
}
