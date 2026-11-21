package server

// === MACHINE-TUNNEL-FORK START ===
// Test SDK presigning without a container image or a running S3 service.
import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/netbirdio/netbird/upload-server/types"
)

func Test_S3HandlerGetUploadURL(t *testing.T) {
	// Presigning happens in the SDK without contacting the bucket.
	var contacted atomic.Bool
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		contacted.Store(true)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(endpoint.Close)

	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ENDPOINT_URL", endpoint.URL)
	t.Setenv("AWS_ENDPOINT_URL_S3", endpoint.URL)
	t.Setenv("AWS_ACCESS_KEY_ID", "test-access-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test-signing-secret")
	t.Setenv("AWS_SESSION_TOKEN", "")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	awsDir := t.TempDir()
	t.Setenv("AWS_CONFIG_FILE", filepath.Join(awsDir, "config"))
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", filepath.Join(awsDir, "credentials"))
	t.Setenv("AWS_PROFILE", "")
	t.Setenv(bucketVar, "test")

	mux := http.NewServeMux()
	require.NoError(t, configureS3Handlers(mux))

	req := httptest.NewRequest(http.MethodGet, types.GetURLPath+"?id=test-file", nil)
	req.Header.Set(types.ClientHeader, types.ClientHeaderValue)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response types.GetURLResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	signed, err := url.Parse(response.URL)
	require.NoError(t, err)
	require.Contains(t, response.Key, "test-file/")
	require.Equal(t, "/test/"+response.Key, signed.Path)
	require.Equal(t, endpoint.URL, signed.Scheme+"://"+signed.Host)
	require.Equal(t, "AWS4-HMAC-SHA256", signed.Query().Get("X-Amz-Algorithm"))
	require.NotEmpty(t, signed.Query().Get("X-Amz-Signature"))
	require.False(t, contacted.Load())
}

// === MACHINE-TUNNEL-FORK END ===
