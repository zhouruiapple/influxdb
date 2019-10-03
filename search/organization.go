package search

import (
	"time"

	"github.com/blevesearch/bleve/document"
	"github.com/influxdata/influxdb"
)

type Organization struct {
	DocType     DocType   `json:"docType"`
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (o Organization) Key() string {
	return o.ID
}

func ConvertOrganization(o influxdb.Organization) Organization {
	return Organization{
		DocType:     DocTypeOrg,
		ID:          o.ID.String(),
		Name:        o.Name,
		Description: o.Description,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}

func newOrganization(d *document.Document) (o Organization) {
	stringMapper := map[string]func(string){
		"docType":     func(s string) { o.DocType = DocType(s) },
		"id":          func(s string) { o.ID = s },
		"name":        func(s string) { o.Name = s },
		"description": func(s string) { o.Description = s },
	}
	for _, f := range d.Fields {
		switch f.Name() {
		case "createdAt":
			tm, ok := getDateTime(f)
			if ok {
				o.CreatedAt = tm
			}
		case "updatedAt":
			tm, ok := getDateTime(f)
			if ok {
				o.UpdatedAt = tm
			}
		default:
			fn, ok := stringMapper[f.Name()]
			if ok {
				fn(string(f.Value()))
			}
		}
	}
	return o
}
