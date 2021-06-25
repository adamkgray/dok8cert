package dok8cert

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/client-go/rest"
)

type credentialsApiResponse struct {
	Server                   string `json:"server"`
	CertificateAuthorityData string `json:"certificate_authority_data"`
	ClientCertificateData    string `json:"client_certificate_data"`
	ClientKeyData            string `json:"client_key_data"`
	Token                    string `json:"token"`
	Expiresat                string `json:"expires_at"`
	Id                       string `json:"id"`
	Message                  string `json:"message"`
}

var apiEndpoint string = "https://api.digitalocean.com/v2/kubernetes/clusters/%s/credentials"

func Update(clusterId string, accessToken string, client *rest.Config) (bool, error) {
	// Call the digitalocean credentials api
	body, err := credentialsApi(clusterId, accessToken)
	if err != nil {
		return false, err
	}

	// Unmarshal the the raw json response into a credentialsApiResponse
	resp, err := unmarshalCredentialsApiResponse(body)
	if err != nil {
		return false, err
	}

	// Handle non-OK responses
	// The digitalocean credentials api will return a non empty string for 'Id' and 'Message' in the case of a non 2XX response
	if resp.Id != "" {
		msg := fmt.Sprintf(
			"digitalocean credentials api response was not 'OK': (%s) %s",
			resp.Id,
			resp.Message,
		)
		return false, errors.New(msg)
	}

	// Decode the base64-encoded cert
	cert, err := decodeCert(resp.CertificateAuthorityData)
	if err != nil {
		return false, err
	}

	// Update the TLS client config in the rest config with the custom cert
	client.TLSClientConfig.CAData = cert

	return true, nil
}

func credentialsApi(clusterId string, accessToken string) ([]byte, error) {
	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf(apiEndpoint, clusterId), nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := httpClient.Do(req)
	if err != nil {
		msg := fmt.Sprintf("failed to call digitalocean credentials api: %s", err)
		return []byte{}, errors.New(msg)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read digitalocean credentials api response body: %s", err)
		return []byte{}, errors.New(msg)
	}
	return body, nil
}

func unmarshalCredentialsApiResponse(body []byte) (credentialsApiResponse, error) {
	data := credentialsApiResponse{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		msg := fmt.Sprintf("failed to unmarshal digitalocean credentials api response json: %s", err)
		return credentialsApiResponse{}, errors.New(msg)
	}
	return data, nil
}

func decodeCert(encodedCert string) ([]byte, error) {
	decodedCert, err := base64.StdEncoding.DecodeString(encodedCert)
	if err != nil {
		msg := fmt.Sprintf("failed to decode TLS certificate: %s", err)
		return []byte{}, errors.New(msg)
	}
	return decodedCert, nil
}
