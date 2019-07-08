package http

import (
	"context"
	"encoding/json"
	"fmt"
	http "net/http"

	influxdb "github.com/influxdata/influxdb"
	platform "github.com/influxdata/influxdb"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

// DeleteBackend is all services and associated parameters required to construct
// the DeleteHandler.
type DeleteBackend struct {
	Logger *zap.Logger

	BucketService       platform.BucketService
	OrganizationService platform.OrganizationService
}

// NewDeleteBackend returns a new instance of DeleteBackend.
func NewDeleteBackend(b *APIBackend) *DeleteBackend {
	return &DeleteBackend{
		Logger: b.Logger.With(zap.String("handler", "delete")),

		BucketService:       b.BucketService,
		OrganizationService: b.OrganizationService,
	}
}

// DeleteHandler receives delete requests.
type DeleteHandler struct {
	*httprouter.Router
	Logger *zap.Logger

	BucketService       platform.BucketService
	OrganizationService platform.OrganizationService
}

const (
	deletePath = "/api/v2/delete"
)

// NewDeleteHandler creates a new handler at /api/v2/delete to delete.
func NewDeleteHandler(b *DeleteBackend) *DeleteHandler {
	h := &DeleteHandler{
		Router: NewRouter(),
		Logger: b.Logger,

		BucketService:       b.BucketService,
		OrganizationService: b.OrganizationService,
	}

	h.HandlerFunc("POST", deletePath, h.handleDelete)
	return h
}

func (h *DeleteHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := decodeDeleteRequest(ctx, r)
	if err != nil {
		EncodeError(ctx, err, w)
		return
	}

	fmt.Printf("delete request body: %v\n", req)

	// h.Logger.log(fmt.Printf("delete request body: %v\n", req))

	// DeletBucketRangePredicate(req.Org, req.Bucket, req.Start, req.Stop, pred tsm1.Predicate)
}

type deleteRequestBody struct {
	Start int64
	Stop  int64
}

type deleteRequest struct {
	Body *deleteRequestBody

	Bucket    influxdb.ID
	Org       influxdb.ID
	Precision string
}

func decodeDeleteRequest(ctx context.Context, r *http.Request) (*deleteRequest, error) {
	qp := r.URL.Query()
	req := &deleteRequest{
		Body: &deleteRequestBody{},
	}
	qp.Get("org")
	qp.Get("bucket")
	qp.Get("precision")

	if org := qp.Get("org"); org != "" {
		id, err := influxdb.IDFromString(org)
		if err != nil {
			return nil, err
		}
		req.Org = *id
	}

	if bucket := qp.Get("bucket"); bucket != "" {
		id, err := influxdb.IDFromString(bucket)
		if err != nil {
			return nil, err
		}
		req.Bucket = *id
	}

	precision := qp.Get("precision")
	if precision == "" {
		precision = "ns"
	}
	req.Precision = precision

	if err := json.NewDecoder(r.Body).Decode(req.Body); err != nil {
		return nil, err
	}

	return req, nil
}

// func (d deleteRequest) Validate() error {
//
// }
