package remotewrite

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/influxdata/influxdb/v2"
	"github.com/influxdata/influxdb/v2/kit/platform"
	"github.com/influxdata/influxdb/v2/replications/metrics"
	replicationsMock "github.com/influxdata/influxdb/v2/replications/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

//go:generate go run github.com/golang/mock/mockgen -package mock -destination ../mock/http_config_store.go github.com/influxdata/influxdb/v2/replications/remotewrite HttpConfigStore

var (
	testID = platform.ID(1)
)

func testWriter(t *testing.T) (*writer, *replicationsMock.MockHttpConfigStore) {
	ctrl := gomock.NewController(t)
	configStore := replicationsMock.NewMockHttpConfigStore(ctrl)
	w := NewWriter(testID, configStore, metrics.NewReplicationsMetrics(), zaptest.NewLogger(t))
	return w, configStore
}

func testServer(t *testing.T, status int, wantData []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotData, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, wantData, gotData)
		w.WriteHeader(status)
	}))
}

func instaWait() waitFunc {
	return func(t time.Duration) <-chan time.Time {
		out := make(chan time.Time)
		close(out)
		return out
	}
}

func TestWrite(t *testing.T) {
	t.Parallel()

	testData := []byte("some data")

	t.Run("error getting config", func(t *testing.T) {
		wantErr := errors.New("uh oh")

		w, configStore := testWriter(t)

		configStore.EXPECT().GetFullHTTPConfig(gomock.Any(), testID).Return(nil, wantErr)
		require.Equal(t, wantErr, w.Write(context.Background(), []byte{}))
	})

	t.Run("nil response from PostWrite", func(t *testing.T) {
		testConfig := &influxdb.ReplicationHTTPConfig{
			RemoteURL: "not a good URL",
		}

		w, configStore := testWriter(t)

		configStore.EXPECT().GetFullHTTPConfig(gomock.Any(), testID).Return(testConfig, nil)
		require.Error(t, w.Write(context.Background(), []byte{}))
	})

	t.Run("immediate good response", func(t *testing.T) {
		svr := testServer(t, http.StatusNoContent, testData)
		defer svr.Close()

		testConfig := &influxdb.ReplicationHTTPConfig{
			RemoteURL: svr.URL,
		}

		w, configStore := testWriter(t)

		configStore.EXPECT().GetFullHTTPConfig(gomock.Any(), testID).Return(testConfig, nil)
		configStore.EXPECT().UpdateResponseInfo(gomock.Any(), testID, http.StatusNoContent, "").Return(nil)
		require.NoError(t, w.Write(context.Background(), testData))
	})

	t.Run("error updating response info", func(t *testing.T) {
		wantErr := errors.New("o no")

		svr := testServer(t, http.StatusNoContent, testData)
		defer svr.Close()

		testConfig := &influxdb.ReplicationHTTPConfig{
			RemoteURL: svr.URL,
		}

		w, configStore := testWriter(t)

		configStore.EXPECT().GetFullHTTPConfig(gomock.Any(), testID).Return(testConfig, nil)
		configStore.EXPECT().UpdateResponseInfo(gomock.Any(), testID, http.StatusNoContent, "").Return(wantErr)
		require.Equal(t, wantErr, w.Write(context.Background(), testData))
	})

	t.Run("bad server responses that never succeed", func(t *testing.T) {
		testAttempts := 3

		for _, status := range []int{http.StatusOK, http.StatusTeapot, http.StatusInternalServerError} {
			t.Run(fmt.Sprintf("status code %d", status), func(t *testing.T) {
				svr := testServer(t, status, testData)
				defer svr.Close()

				testConfig := &influxdb.ReplicationHTTPConfig{
					RemoteURL: svr.URL,
				}

				w, configStore := testWriter(t)
				w.waitFunc = instaWait()
				w.maximumAttemptsBeforeErr = testAttempts

				configStore.EXPECT().GetFullHTTPConfig(gomock.Any(), testID).Return(testConfig, nil).Times(testAttempts)
				configStore.EXPECT().UpdateResponseInfo(gomock.Any(), testID, status, invalidResponseCode(status).Error()).Return(nil).Times(testAttempts)
				require.Equal(t, errors.New("maximum number of attempts exceeded"), w.Write(context.Background(), testData))
			})
		}
	})

	t.Run("bad server responses at first followed by good server responses", func(t *testing.T) {
		testAttempts := 10
		attemptsBeforeSuccess := 3
		serverCounter := 0
		badStatus := http.StatusInternalServerError
		goodStatus := http.StatusNoContent

		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serverCounter++
			gotData, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)
			require.Equal(t, testData, gotData)
			if serverCounter >= attemptsBeforeSuccess {
				w.WriteHeader(goodStatus)
				return
			}

			w.WriteHeader(badStatus)
		}))
		defer svr.Close()

		testConfig := &influxdb.ReplicationHTTPConfig{
			RemoteURL: svr.URL,
		}

		w, configStore := testWriter(t)
		w.waitFunc = instaWait()
		w.maximumAttemptsBeforeErr = testAttempts

		configStore.EXPECT().GetFullHTTPConfig(gomock.Any(), testID).Return(testConfig, nil).Times(attemptsBeforeSuccess)
		configStore.EXPECT().UpdateResponseInfo(gomock.Any(), testID, badStatus, invalidResponseCode(badStatus).Error()).Return(nil).Times(attemptsBeforeSuccess - 1)
		configStore.EXPECT().UpdateResponseInfo(gomock.Any(), testID, goodStatus, "").Return(nil)
		require.NoError(t, w.Write(context.Background(), testData))
	})

	t.Run("client timeouts", func(t *testing.T) {
		attemptsBeforeSuccess := 2
		serverCounter := 0
		serverDelay := time.Second
		clientTimeout := time.Millisecond

		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serverCounter++
			gotData, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)
			require.Equal(t, testData, gotData)

			if serverCounter < attemptsBeforeSuccess {
				time.Sleep(serverDelay)
			}

			w.WriteHeader(http.StatusNoContent)
		}))
		defer svr.Close()

		testConfig := &influxdb.ReplicationHTTPConfig{
			RemoteURL: svr.URL,
		}

		w, configStore := testWriter(t)
		w.waitFunc = instaWait()
		w.clientTimeout = clientTimeout

		configStore.EXPECT().GetFullHTTPConfig(gomock.Any(), testID).Return(testConfig, nil).Times(attemptsBeforeSuccess)
		configStore.EXPECT().UpdateResponseInfo(gomock.Any(), testID, 0, gomock.Any()).Return(nil).Times(attemptsBeforeSuccess - 1)
		configStore.EXPECT().UpdateResponseInfo(gomock.Any(), testID, http.StatusNoContent, "").Return(nil)
		require.NoError(t, w.Write(context.Background(), testData))
	})

	t.Run("drops bad data after config is updated", func(t *testing.T) {
		testAttempts := 5

		svr := testServer(t, http.StatusBadRequest, testData)
		defer svr.Close()

		testConfig := &influxdb.ReplicationHTTPConfig{
			RemoteURL: svr.URL,
		}

		updatedConfig := &influxdb.ReplicationHTTPConfig{
			RemoteURL:            svr.URL,
			DropNonRetryableData: true,
		}

		w, configStore := testWriter(t)
		w.waitFunc = instaWait()
		w.maximumAttemptsBeforeErr = testAttempts

		configStore.EXPECT().GetFullHTTPConfig(gomock.Any(), testID).Return(testConfig, nil).Times(testAttempts - 1)
		configStore.EXPECT().GetFullHTTPConfig(gomock.Any(), testID).Return(updatedConfig, nil)
		configStore.EXPECT().UpdateResponseInfo(gomock.Any(), testID, http.StatusBadRequest, invalidResponseCode(http.StatusBadRequest).Error()).Return(nil).Times(testAttempts)
		require.NoError(t, w.Write(context.Background(), testData))
	})

	t.Run("uses wait time from response header if present", func(t *testing.T) {
		testAttempts := 3
		numSeconds := 5
		waitTimeFromHeader := 5 * time.Second

		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotData, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)
			require.Equal(t, testData, gotData)
			w.Header().Set(retryAfterHeaderKey, strconv.Itoa(numSeconds))
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer svr.Close()

		testConfig := &influxdb.ReplicationHTTPConfig{
			RemoteURL: svr.URL,
		}

		w, configStore := testWriter(t)
		w.waitFunc = func(dur time.Duration) <-chan time.Time {
			require.Equal(t, waitTimeFromHeader, dur)
			out := make(chan time.Time)
			close(out)
			return out
		}
		w.maximumAttemptsBeforeErr = testAttempts

		configStore.EXPECT().GetFullHTTPConfig(gomock.Any(), testID).Return(testConfig, nil).Times(testAttempts)
		configStore.EXPECT().UpdateResponseInfo(gomock.Any(), testID, http.StatusTooManyRequests, invalidResponseCode(http.StatusTooManyRequests).Error()).Return(nil).Times(testAttempts)
		require.Equal(t, errors.New("maximum number of attempts exceeded"), w.Write(context.Background(), testData))
	})

	t.Run("can cancel with context", func(t *testing.T) {
		svr := testServer(t, http.StatusInternalServerError, testData)
		defer svr.Close()

		testConfig := &influxdb.ReplicationHTTPConfig{
			RemoteURL: svr.URL,
		}

		ctx, cancel := context.WithCancel(context.Background())

		w, configStore := testWriter(t)
		w.waitFunc = func(dur time.Duration) <-chan time.Time {
			cancel()
			return time.After(time.Second)
		}

		configStore.EXPECT().GetFullHTTPConfig(gomock.Any(), testID).Return(testConfig, nil)
		configStore.EXPECT().UpdateResponseInfo(gomock.Any(), testID, http.StatusInternalServerError, invalidResponseCode(http.StatusInternalServerError).Error()).Return(nil)
		require.Equal(t, context.Canceled, w.Write(ctx, testData))
	})
}

func TestPostWrite(t *testing.T) {
	testData := []byte("some data")

	tests := []struct {
		status  int
		wantErr bool
	}{
		{
			status:  http.StatusOK,
			wantErr: true,
		},
		{
			status:  http.StatusNoContent,
			wantErr: false,
		},
		{
			status:  http.StatusBadRequest,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status code %d", tt.status), func(t *testing.T) {
			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				recData, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, testData, recData)

				w.WriteHeader(tt.status)
			}))
			defer svr.Close()

			config := &influxdb.ReplicationHTTPConfig{
				RemoteURL: svr.URL,
			}

			res, err := PostWrite(context.Background(), config, testData, time.Second)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.Nil(t, err)
			}

			require.Equal(t, tt.status, res.StatusCode)
		})
	}
}

func TestWaitTimeFromHeader(t *testing.T) {
	w := &writer{
		maximumAttemptsForBackoffTime: maximumAttempts,
	}

	tests := []struct {
		headerKey string
		headerVal string
		want      time.Duration
	}{
		{
			headerKey: retryAfterHeaderKey,
			headerVal: "30",
			want:      30 * time.Second,
		},
		{
			headerKey: retryAfterHeaderKey,
			headerVal: "0",
			want:      w.backoff(1),
		},
		{
			headerKey: retryAfterHeaderKey,
			headerVal: "not a number",
			want:      0,
		},
		{
			headerKey: "some other thing",
			headerVal: "not a number",
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q - %q", tt.headerKey, tt.headerVal), func(t *testing.T) {
			r := &http.Response{
				Header: http.Header{
					tt.headerKey: []string{tt.headerVal},
				},
			}

			got := w.waitTimeFromHeader(r)
			require.Equal(t, tt.want, got)
		})
	}
}
