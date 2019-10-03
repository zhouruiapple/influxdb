package search

import (
	"github.com/blevesearch/bleve/document"
	"github.com/influxdata/influxdb"
)

// ConvertUser will convert a domain user to a search.User.
func ConvertUser(u influxdb.User) User {
	return User{
		DocType: DocTypeUser,
		ID:      u.ID.String(),
		Name:    u.Name,
		OAuthID: u.OAuthID,
		Status:  u.Status,
	}
}

// newUser create bucket from bleve fields.
func newUser(d *document.Document) (u User) {
	stringMapper := map[string]func(string){
		"docType": func(s string) { u.DocType = DocType(s) },
		"id":      func(s string) { u.ID = s },
		"name":    func(s string) { u.Name = s },
		"oauthID": func(s string) { u.OAuthID = s },
		"status":  func(s string) { u.Status = influxdb.Status(s) },
	}
	for _, f := range d.Fields {
		switch f.Name() {
		default:
			fn, ok := stringMapper[f.Name()]
			if ok {
				fn(string(f.Value()))
			}
		}
	}

	return u
}

// User implements Doc interface.
type User struct {
	DocType DocType         `json:"docType"`
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	OAuthID string          `json:"oauthID,omitempty"`
	Status  influxdb.Status `json:"status"`
}

// Key returns the primary key.
func (u User) Key() string {
	return u.ID
}
