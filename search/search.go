package search

import (
	"context"
	"fmt"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
	"github.com/blevesearch/bleve/search/query"
	"github.com/influxdata/influxdb"
	"go.uber.org/zap"
)

// Doc the document for search.
type Doc interface {
	Key() string
}

// FindService is the search service.
type FindService interface {
	SimpleQuery(q string, dt DocType) ([]Doc, error)
	Index(ctx context.Context, doc Doc) error
}

// serivce is the service of search.
type service struct {
	Core   bleve.Index
	Logger *zap.Logger
}

// NewService is constructor of the search service.
func NewService(core bleve.Index, logger *zap.Logger) FindService {
	return &service{
		Core:   core,
		Logger: logger,
	}
}

// SimpleQuery search for a simple text.
func (s *service) SimpleQuery(q string, dt DocType) ([]Doc, error) {
	queries := []query.Query{
		bleve.NewFuzzyQuery(q),
	}
	if dt != DocTypeUnknown {
		queries = append(queries, bleve.NewMatchQuery(fmt.Sprintf(`+DocType:%q`, dt)))
	}

	req := bleve.NewSearchRequest(bleve.NewConjunctionQuery(queries...))
	results, err := s.Core.Search(req)
	if err != nil {
		return nil, err
	}

	docs := make([]Doc, 0, len(results.Hits))
	for _, h := range results.Hits {
		doc, err := s.Core.Document(h.ID)
		if err != nil || doc == nil {
			continue
		}

		docuType := getDocType(doc.Fields)
		switch docuType {
		case DocTypeBucket:
			docs = append(docs, newBucket(doc))
		case DocTypeOrg:
			docs = append(docs, newOrganization(doc))
		case DocTypeUser:
			docs = append(docs, newUser(doc))
		}
	}

	return docs, nil
}

func getDocType(fields []document.Field) DocType {
	for _, f := range fields {
		if f.Name() == "docType" {
			return DocType(string(f.Value()))
		}
	}
	return DocTypeUnknown
}

// Index a document with key.
func (s *service) Index(ctx context.Context, doc Doc) error {
	return s.Core.Index(doc.Key(), doc)
}

// DocType is an Enum type of search type.
type DocType string

// doc types
const (
	DocTypeUnknown DocType = ""
	DocTypeBucket  DocType = "bucket"
	DocTypeOrg     DocType = "org"
	DocTypeUser    DocType = "user"
)

// ConvertBucket will convert a domain bucket to a search.Bucket.
func ConvertBucket(b influxdb.Bucket) Bucket {
	return Bucket{
		DocType:             DocTypeBucket,
		ID:                  b.ID.String(),
		OrgID:               b.OrgID.String(),
		Type:                b.Type.String(),
		Name:                b.Name,
		Description:         b.Description,
		RetentionPeriod:     b.RetentionPeriod.String(),
		RetentionPolicyName: b.RetentionPolicyName,
		CreatedAt:           b.CreatedAt,
		UpdatedAt:           b.UpdatedAt,
	}
}

// newBucket create bucket from bleve fields.
func newBucket(d *document.Document) (b Bucket) {
	stringMapper := map[string]func(string){
		"docType":             func(s string) { b.DocType = DocType(s) },
		"id":                  func(s string) { b.ID = s },
		"orgID":               func(s string) { b.OrgID = s },
		"type":                func(s string) { b.Type = s },
		"name":                func(s string) { b.Name = s },
		"description":         func(s string) { b.Description = s },
		"retentionPeriod":     func(s string) { b.RetentionPeriod = s },
		"retentionPolicyName": func(s string) { b.RetentionPolicyName = s },
	}
	for _, f := range d.Fields {
		switch f.Name() {
		case "createdAt":
			tm, ok := getDateTime(f)
			if ok {
				b.CreatedAt = tm
			}
		case "updatedAt":
			tm, ok := getDateTime(f)
			if ok {
				b.UpdatedAt = tm
			}
		default:
			fn, ok := stringMapper[f.Name()]
			if ok {
				fn(string(f.Value()))
			}
		}
	}

	return b
}

func getDateTime(f document.Field) (time.Time, bool) {
	tf := document.NewDateTimeFieldFromBytes(f.Name(), f.ArrayPositions(), f.Value())
	tm, err := tf.DateTime()
	return tm, err == nil
}

// Bucket implements Doc interface.
type Bucket struct {
	DocType             DocType   `json:"docType"`
	ID                  string    `json:"id"`
	OrgID               string    `json:"orgID"`
	Type                string    `json:"type"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	RetentionPeriod     string    `json:"retentionPeriod"`
	RetentionPolicyName string    `json:"retentionPolicyName"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// Key returns the primary key.
func (b Bucket) Key() string {
	return b.ID
}
