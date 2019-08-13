package http_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxdb/http"
)

func TestEncodeError(t *testing.T) {
	ctx := context.TODO()

	w := httptest.NewRecorder()

	http.ErrorHandler(0).HandleHTTPError(ctx, nil, w)

	if w.Code != 200 {
		t.Errorf("expected status code 200, got: %d", w.Code)
	}
}

func TestEncodeErrorWithError(t *testing.T) {
	ctx := context.TODO()
	err := fmt.Errorf("there's an error here, be aware")

	w := httptest.NewRecorder()

	http.ErrorHandler(0).HandleHTTPError(ctx, err, w)

	if w.Code != 500 {
		t.Errorf("expected status code 500, got: %d", w.Code)
	}

	errHeader := w.Header().Get("X-Platform-Error-Code")
	if errHeader != influxdb.EInternal {
		t.Errorf("expected X-Platform-Error-Code: %s, got: %s", influxdb.EInternal, errHeader)
	}

	expected := &influxdb.Error{
		Code: influxdb.EInternal,
		Err:  err,
	}

	pe := http.CheckError(w.Result())
	if pe.(*influxdb.Error).Err.Error() != expected.Err.Error() {
		t.Errorf("errors encode err: got %s", w.Body.String())
	}
}

func TestCheckErrorContentType(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		body        []byte
		contentType string
		code        int
		// expecations
		errCode    string
		errMessage string
	}{
		{
			name:        "text/plain status code 500",
			body:        []byte("something went wrong"),
			contentType: "text/plain",
			code:        500,
			errCode:     influxdb.EInternal,
			errMessage:  "something went wrong",
		},
		{
			name:        "blank content type status code 504",
			body:        []byte("upstream request timeout"),
			contentType: "",
			code:        504,
			errCode:     influxdb.EInternal,
			errMessage:  "upstream request timeout",
		},
		{
			name:        "application/json; charset=utf-8 status code 404",
			body:        []byte(`{"code":"not found", "error": "could not find resource"}`),
			contentType: "application/json; charset=utf-8",
			code:        404,
			errCode:     influxdb.ENotFound,
			errMessage:  "could not find resource",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			w.Header().Set("Content-Type", testCase.contentType)
			w.WriteHeader(testCase.code)
			w.Write(testCase.body)

			err := http.CheckError(w.Result())
			if err == nil {
				t.Error("err must not be nil")
			}

			ierr, ok := err.(*influxdb.Error)
			if !ok {
				t.Errorf("should be *influxdb.Error, got: %q", err)
				return
			}

			if testCase.errCode != ierr.Code {
				t.Errorf("expected %q, got %q", testCase.errCode, ierr.Code)
				return
			}

			if testCase.errMessage != ierr.Err.Error() {
				t.Errorf("expected %q, got %q", testCase.errMessage, ierr.Err)
			}
		})
	}
}
