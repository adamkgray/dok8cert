package dok8cert

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"k8s.io/client-go/rest"
)

type mockNetworkClient struct {
	fail bool
	body string
}

func (mnc mockNetworkClient) Do(req *http.Request) (*http.Response, error) {
	if mnc.fail {
		return &http.Response{}, errors.New("fail")
	}
	var mockBodyReader io.Reader = strings.NewReader(mnc.body)

	var mockBodyReadCloser io.ReadCloser = io.NopCloser(mockBodyReader)
	return &http.Response{
		Body: mockBodyReadCloser,
	}, nil
}

func TestDecodeCert(t *testing.T) {
	cases := []struct {
		desc         string
		encodedCert  string
		expectedCert []byte
		expectedErr  error
	}{
		{
			"TestValidCert",
			"SGVsbG8sIHdvcmxkLg==",
			[]byte("Hello, world."),
			nil,
		},
		{
			"TestInvalidCert",
			"Hello, world.",
			[]byte{},
			errors.New("failed to decode TLS certificate: illegal base64 data at input byte 5"),
		},
	}

	for _, tc := range cases {
		actualCert, actualErr := decodeCert(tc.encodedCert)

		// byte slices cannot be compared outright,
		// so instead we compare the length and content
		// compare length first to avoid index out of range panic
		if len(tc.expectedCert) != len(actualCert) {
			t.Fatalf("%s: expectedCert: %s got: %s for encodedCert: %s", tc.desc, tc.expectedErr, actualErr, tc.encodedCert)
		}
		for i := range actualCert {
			if actualCert[i] != tc.expectedCert[i] {
				t.Fatalf("%s: expectedCert: %s got: %s for encodedCert: %s", tc.desc, tc.expectedErr, actualErr, tc.encodedCert)
				break
			}
		}

		// nil error case
		if (tc.expectedErr == nil || actualErr == nil) && (tc.expectedErr != actualErr) {
			t.Fatalf("%s: expectedErr: %s got: %s for encodedCert: %s", tc.desc, tc.expectedErr, actualErr, tc.encodedCert)
		}
		// actual error case
		if (tc.expectedErr != nil && actualErr != nil) && (tc.expectedErr.Error() != actualErr.Error()) {
			t.Fatalf("%s: expectedErr: %s got: %s for encodedCert: %s", tc.desc, tc.expectedErr, actualErr, tc.encodedCert)
		}

	}
}
func BenchmarkDecodeCert(b *testing.B) {
	encodedCert := "SGVsbG8sIHdvcmxkLg=="
	for i := 0; i < b.N; i++ {
		decodeCert(encodedCert)
	}
}
func TestUnmarshalCredentialsApiResponse(t *testing.T) {
	cases := []struct {
		desc         string
		body         []byte
		expectedResp credentialsApiResponse
		expectedErr  error
	}{
		{
			"TestOkayBody",
			[]byte("{\"server\": \"https://cluster-id.k8s.ondigitalocean.com\",\"certificate_authority_data\": \"SGVsbG8sIHdvcmxkLg==\",\"client_certificate_data\": null,\"client_key_data\": null,\"token\": \"token\",\"expires_at\": \"YYYY-MM-DDThh:mm:ssZ\"}"),
			credentialsApiResponse{
				Server:                   "https://cluster-id.k8s.ondigitalocean.com",
				CertificateAuthorityData: "SGVsbG8sIHdvcmxkLg==",
				ClientCertificateData:    "",
				ClientKeyData:            "",
				Token:                    "token",
				ExpiresAt:                "YYYY-MM-DDThh:mm:ssZ",
			},
			nil,
		},
		{
			"TestErrorBody",
			[]byte("{\"id\": \"id\", \"message\": \"message\"}"),
			credentialsApiResponse{
				Id:      "id",
				Message: "message",
			},
			nil,
		},
		{
			"TestInvalidBody",
			[]byte("invalid json"),
			credentialsApiResponse{},
			errors.New("failed to unmarshal digitalocean credentials api response json: invalid character 'i' looking for beginning of value"),
		},
	}

	for _, tc := range cases {
		actualResp, actualErr := unmarshalCredentialsApiResponse(tc.body)

		// check returned struct
		if tc.expectedResp != actualResp {
			t.Fatalf("%s: expectedResp: %+v got: %+v for body: %s", tc.desc, tc.expectedResp, actualResp, tc.body)
		}

		// nil error case
		if (tc.expectedErr == nil || actualErr == nil) && (tc.expectedErr != actualErr) {
			t.Fatalf("%s: expectedErr: %s got: %s for body: %s", tc.desc, tc.expectedErr, actualErr, tc.body)
		}
		// actual error case
		if (tc.expectedErr != nil && actualErr != nil) && (tc.expectedErr.Error() != actualErr.Error()) {
			t.Fatalf("%s: expectedErr: %s got: %s for body: %s", tc.desc, tc.expectedErr, actualErr, tc.body)
		}
	}
}
func BenchmarkUnmarshalCredentialsApiResponse(b *testing.B) {
	body := []byte("{\"server\": \"https://cluster-id.k8s.ondigitalocean.com\",\"certificate_authority_data\": \"SGVsbG8sIHdvcmxkLg==\",\"client_certificate_data\": null,\"client_key_data\": null,\"token\": \"token\",\"expires_at\": \"YYYY-MM-DDThh:mm:ssZ\"}")
	for i := 0; i < b.N; i++ {
		unmarshalCredentialsApiResponse(body)
	}
}
func TestOK(t *testing.T) {
	cases := []struct {
		desc        string
		resp        credentialsApiResponse
		expectedOK  bool
		expectedErr error
	}{
		{
			"TestOkay",
			credentialsApiResponse{
				Server:                   "https://cluster-id.k8s.ondigitalocean.com",
				CertificateAuthorityData: "SGVsbG8sIHdvcmxkLg==",
				ClientCertificateData:    "",
				ClientKeyData:            "",
				Token:                    "token",
				ExpiresAt:                "YYYY-MM-DDThh:mm:ssZ",
			},
			true,
			nil,
		},
		{
			"TestNotOkay",
			credentialsApiResponse{
				Id:      "id",
				Message: "message",
			},
			false,
			errors.New("digitalocean credentials api response was not 'OK': (id) message"),
		},
	}

	for _, tc := range cases {
		actualOK, actualErr := tc.resp.OK()
		if tc.expectedOK != actualOK {
			t.Fatalf("%s: expectedOK: %s got: %s for resp: %s", tc.desc, tc.expectedErr, actualErr, tc.resp)
		}

		// nil error case
		if (tc.expectedErr == nil || actualErr == nil) && (tc.expectedErr != actualErr) {
			t.Fatalf("%s: expectedErr: %s got: %s for body: %+v", tc.desc, tc.expectedErr, actualErr, tc.resp)
		}
		// actual error case
		if (tc.expectedErr != nil && actualErr != nil) && (tc.expectedErr.Error() != actualErr.Error()) {
			t.Fatalf("%s: expectedErr: %s got: %s for body: %+v", tc.desc, tc.expectedErr, actualErr, tc.resp)
		}

	}
}
func BenchmarkOK(b *testing.B) {
	resp := credentialsApiResponse{}
	for i := 0; i < b.N; i++ {
		resp.OK()
	}
}

func TestCredentialsApi(t *testing.T) {
	cases := []struct {
		desc          string
		networkClient networkClient
		expectedBody  []byte
		expectedErr   error
	}{
		{
			"TestCredentialsApi",
			mockNetworkClient{
				body: "{\"server\": \"https://cluster-id.k8s.ondigitalocean.com\",\"certificate_authority_data\": \"SGVsbG8sIHdvcmxkLg==\",\"client_certificate_data\": null,\"client_key_data\": null,\"token\": \"token\",\"expires_at\": \"YYYY-MM-DDThh:mm:ssZ\"}",
			},
			[]byte("{\"server\": \"https://cluster-id.k8s.ondigitalocean.com\",\"certificate_authority_data\": \"SGVsbG8sIHdvcmxkLg==\",\"client_certificate_data\": null,\"client_key_data\": null,\"token\": \"token\",\"expires_at\": \"YYYY-MM-DDThh:mm:ssZ\"}"),
			nil,
		},
		{
			"TestCredentialsApiFail",
			mockNetworkClient{
				fail: true,
			},
			[]byte(""),
			errors.New("failed to call digitalocean credentials api: fail"),
		},
	}

	for _, tc := range cases {
		actualBody, actualErr := credentialsApi(tc.networkClient, "clusterId", "accessToken")

		if string(tc.expectedBody) != string(actualBody) {
			t.Fatalf("%s: expectedBody: %s got: %s for networkClient: %+v", tc.desc, string(tc.expectedBody), string(actualBody), tc.networkClient)
		}

		// nil error case
		if (tc.expectedErr == nil || actualErr == nil) && (tc.expectedErr != actualErr) {
			t.Fatalf("%s: expectedErr: %s got: %s for networkClient: %+v", tc.desc, tc.expectedErr, actualErr, tc.networkClient)
		}
		// actual error case
		if (tc.expectedErr != nil && actualErr != nil) && (tc.expectedErr.Error() != actualErr.Error()) {
			t.Fatalf("%s: expectedErr: %s got: %s for networkClient: %+v", tc.desc, tc.expectedErr, actualErr, tc.networkClient)
		}

	}
}
func BenchmarkCredentialsApi(b *testing.B) {
	mnc := mockNetworkClient{
		body: "{\"server\": \"https://cluster-id.k8s.ondigitalocean.com\",\"certificate_authority_data\": \"SGVsbG8sIHdvcmxkLg==\",\"client_certificate_data\": null,\"client_key_data\": null,\"token\": \"token\",\"expires_at\": \"YYYY-MM-DDThh:mm:ssZ\"}",
	}
	for i := 0; i < b.N; i++ {
		credentialsApi(mnc, "clusterId", "accessToken")
	}
}

func TestUpdate(t *testing.T) {
	cases := []struct {
		desc           string
		k8sClient      *rest.Config
		networkClient  networkClient
		expectedOK     bool
		expectedErr    error
		expectedCAData string
	}{
		{
			"TestUpdate",
			&rest.Config{},
			mockNetworkClient{
				body: "{\"server\": \"https://cluster-id.k8s.ondigitalocean.com\",\"certificate_authority_data\": \"SGVsbG8sIHdvcmxkLg==\",\"client_certificate_data\": null,\"client_key_data\": null,\"token\": \"token\",\"expires_at\": \"YYYY-MM-DDThh:mm:ssZ\"}",
			},
			true,
			nil,
			"Hello, world.",
		},
		{
			"TestUpdateFailCredentialsApi",
			&rest.Config{},
			mockNetworkClient{
				fail: true,
				body: "",
			},
			false,
			errors.New("failed to call digitalocean credentials api: fail"),
			"",
		},
		{
			"TestUpdateFailUnmarshalCredentialsApiResponse",
			&rest.Config{},
			mockNetworkClient{
				body: "invalid",
			},
			false,
			errors.New("failed to unmarshal digitalocean credentials api response json: invalid character 'i' looking for beginning of value"),
			"",
		},
		{
			"TestUpdateFailOK",
			&rest.Config{},
			mockNetworkClient{
				body: "{\"id\": \"id\", \"message\": \"message\"}",
			},
			false,
			errors.New("digitalocean credentials api response was not 'OK': (id) message"),
			"",
		},
		{
			"TestUpdateFailDecodeCert",
			&rest.Config{},
			mockNetworkClient{
				body: "{\"certificate_authority_data\": \"Hello, world.\"}",
			},
			false,
			errors.New("failed to decode TLS certificate: illegal base64 data at input byte 5"),
			"",
		},
	}

	for _, tc := range cases {
		actualOK, actualErr := update("cluserId", "accessToken", tc.k8sClient, tc.networkClient)

		if tc.expectedOK != actualOK {
			t.Fatalf("%s: expectedOK: %s got: %s for networkClient: %+v", tc.desc, tc.expectedErr, actualErr, tc.networkClient)
		}

		if tc.expectedCAData != string(tc.k8sClient.TLSClientConfig.CAData) {
			t.Fatalf("%s: expectedCADAta: %s got: %s for networkClient: %+v", tc.desc, tc.expectedCAData, string(tc.k8sClient.TLSClientConfig.CAData), tc.networkClient)
		}

		// nil error case
		if (tc.expectedErr == nil || actualErr == nil) && (tc.expectedErr != actualErr) {
			t.Fatalf("%s: expectedErr: %s got: %s for networkClient: %+v", tc.desc, tc.expectedErr, actualErr, tc.networkClient)
		}
		// actual error case
		if (tc.expectedErr != nil && actualErr != nil) && (tc.expectedErr.Error() != actualErr.Error()) {
			t.Fatalf("%s: expectedErr: %s got: %s for networkClient: %+v", tc.desc, tc.expectedErr, actualErr, tc.networkClient)
		}
	}
}

func BenchmarkUpdate(b *testing.B) {
	mnc := mockNetworkClient{
		body: "{\"server\": \"https://cluster-id.k8s.ondigitalocean.com\",\"certificate_authority_data\": \"SGVsbG8sIHdvcmxkLg==\",\"client_certificate_data\": null,\"client_key_data\": null,\"token\": \"token\",\"expires_at\": \"YYYY-MM-DDThh:mm:ssZ\"}",
	}
	k8sClient := &rest.Config{}
	for i := 0; i < b.N; i++ {
		update("clusterId", "accessToken", k8sClient, mnc)
	}
}
