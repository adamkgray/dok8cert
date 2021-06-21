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
}

var digitalOceanApi string = "https://api.digitalocean.com/v2/kubernetes/clusters/%s/credentials"

func Update(clusterId string, accessToken string, client *rest.Config) (bool, error) {
	// get json response from credentials api
	body, err := credentialsApi(clusterId, accessToken)
	if err != nil {
		return false, err
	}

	// unmarshal response
	resp, err := unmarshalCredentialsApiResponse(body)
	if err != nil {
		return false, err
	}

	// decode cert
	cert, err := decodeCert(resp.CertificateAuthorityData)
	if err != nil {
		return false, err
	}

	// update cert
	client.TLSClientConfig.CAData = cert

	return true, nil
}

func credentialsApi(clusterId string, accessToken string) ([]byte, error) {
	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf(digitalOceanApi, clusterId), nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := httpClient.Do(req)
	if err != nil {
		msg := fmt.Sprintf("failed to call digitalocean credentials api: %s", err)
		return []byte{}, errors.New(msg)
	}
	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("digitalocean credentials api responded with non-2XX HTTP status %d", resp.StatusCode)
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
