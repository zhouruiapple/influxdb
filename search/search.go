package search

import (
	"context"
	"time"

	"github.com/blevesearch/bleve/search/query"

	"github.com/blevesearch/bleve/document"

	"github.com/blevesearch/bleve"

	"github.com/influxdata/influxdb"
)

/*
	way to import the index
	way to search the created index
		*
	way to update the index on every req
*/

type Search interface {
	Search()
}

type Query interface {
	Query()
}

type Doc interface {
	Key() string
}

// Service is the service of search.
type Service struct {
	Core bleve.Index
}

type SearchFilter struct {
	IndexType string
}

// SimpleQuery search for a simple text.
func (s *Service) SimpleQuery(q string, dt DocType) ([]Doc, error) {
	queries := []query.Query{bleve.NewQueryStringQuery(q)}
	if dt != DocTypeUnknown {
		bleve.New
		queries = append(queries, bleve.NewTermQuery(`IndexType:"bucket"`))
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
		default:
			continue
		}
	}

	return docs, nil
}

func getDocType(fields []document.Field) DocType {
	for _, f := range fields {
		if f.Name() == "DocType" {
			return DocType(string(f.Value()))
		}
	}
	return DocTypeUnknown
}

// Index a document with key.
func (s *Service) Index(ctx context.Context, doc Doc) error {
	return s.Core.Index(doc.Key(), doc)
}

type DocType string

const (
	DocTypeUnknown DocType = ""
	DocTypeBucket  DocType = "bucket"
)

func ConvertBucket(b influxdb.Bucket) Bucket {
	return Bucket{
		IndexType:           DocTypeBucket,
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
		"DocType":             func(s string) { b.IndexType = DocType(s) },
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
			tf := document.NewDateTimeFieldFromBytes(f.Name(), f.ArrayPositions(), f.Value())
			tm, err := tf.DateTime()
			if err == nil {
				b.CreatedAt = tm
			}
		case "updatedAt":
			tf := document.NewDateTimeFieldFromBytes(f.Name(), f.ArrayPositions(), f.Value())
			tm, err := tf.DateTime()
			if err == nil {
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

// Bucket implements Doc interface.
type Bucket struct {
	IndexType           DocType   `json:"DocType"`
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
