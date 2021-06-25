package dok8cert

import (
	"testing"
)

func TestDecodeCert(t *testing.T) {
	// happy
	encodedCert := "SGVsbG8sIHdvcmxkLg=="
	wanted := []byte("Hello, world.")
	got, _ := decodeCert(encodedCert)
	if len(wanted) != len(got) {
		t.Errorf("failed correctly decode cert: wanted %s, got %s", wanted, got)
	}
	for i := range got {
		if got[i] != wanted[i] {
			t.Errorf("failed correctly decode cert: wanted %s, got %s", wanted, got)
		}
	}

	// bad
	invalidEncodedCert := "Hello, world."
	_, err := decodeCert(invalidEncodedCert)
	if err == nil {
		t.Errorf("invalid base64 string should return error")
	}
}

func BenchmarkDecodeCert(b *testing.B) {
	encodedCert := "SGVsbG8sIHdvcmxkLg=="
	for i := 0; i < b.N; i++ {
		decodeCert(encodedCert)
	}
}

func TestUnmarshalCredentialsApiResponse(t *testing.T) {
	// happy
	okayBody := []byte("{\"server\": \"https://cluster-id.k8s.ondigitalocean.com\",\"certificate_authority_data\": \"SGVsbG8sIHdvcmxkLg==\",\"client_certificate_data\": null,\"client_key_data\": null,\"token\": \"token\",\"expires_at\": \"YYYY-MM-DDThh:mm:ssZ\"}")
	okayResp, _ := unmarshalCredentialsApiResponse(okayBody)
	if okayResp.Server != "https://cluster-id.k8s.ondigitalocean.com" {
		t.Error("failed to coerce json server to struct")
	}
	if okayResp.CertificateAuthorityData != "SGVsbG8sIHdvcmxkLg==" {
		t.Error("failed to coerce json certificate_authority_data to struct")
	}
	if okayResp.ClientCertificateData != "" {
		t.Error("failed to coerce json client_certificate_data to struct")
	}
	if okayResp.ClientKeyData != "" {
		t.Error("failed to coerce json client_key_data to struct")
	}
	if okayResp.Token != "token" {
		t.Error("failed to coerce json token to struct")
	}
	if okayResp.ExpiresAt != "YYYY-MM-DDThh:mm:ssZ" {
		t.Error("failed to coerce json expires_at to struct")
	}

	failedBody := []byte("{\"id\": \"id\", \"message\": \"message\"}")
	failedResp, _ := unmarshalCredentialsApiResponse(failedBody)
	if failedResp.Id != "id" {
		t.Error("failed to coerce json id to struct")
	}
	if failedResp.Message != "message" {
		t.Error("failed to coerce json message to struct")
	}

	// bad
	invalidBody := []byte("invalid json")
	_, err := unmarshalCredentialsApiResponse(invalidBody)
	if err == nil {
		t.Error("invalid json should return error")
	}
}

func BenchmarkUnmarshalCredentialsApiResponse(b *testing.B) {
	okayBody := []byte("{\"server\": \"https://cluster-id.k8s.ondigitalocean.com\",\"certificate_authority_data\": \"SGVsbG8sIHdvcmxkLg==\",\"client_certificate_data\": null,\"client_key_data\": null,\"token\": \"token\",\"expires_at\": \"YYYY-MM-DDThh:mm:ssZ\"}")
	for i := 0; i < b.N; i++ {
		unmarshalCredentialsApiResponse(okayBody)
	}
}
func TestCredentialsApi(t *testing.T) {
	// happy

	// bad
}

func TestOK(t *testing.T) {
	// happy
	resp := credentialsApiResponse{}
	ok, err := resp.OK()
	if !ok {
		t.Errorf("expected okay result, got not okay: %s", err)
	}

	// bad
	resp.Id = "not ok"
	resp.Message = "not ok"
	ok, _ = resp.OK()
	if ok {
		t.Errorf("expected not okay result, got okay: id=%s, message=%s", resp.Id, resp.Message)
	}
}

func BenchmarkOK(b *testing.B) {
	resp := credentialsApiResponse{}
	for i := 0; i < b.N; i++ {
		resp.OK()
	}
}
