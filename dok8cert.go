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

var digitalOceanApi string = "https://api.digitalocean.com/v2/kubernetes/clusters/%s/credentials"

func Set(cert []byte, client *rest.Config) {
	client.TLSClientConfig.CAData = cert
}

func Get(clusterId string, accessToken string) ([]byte, error) {
	// call DO credentials Api
	resp, err := credentialsApi(clusterId, accessToken)
	if err != nil {
		return []byte{}, err
	}

	// parse response
	data, err := parseCredentialsApiResponse(resp)
	if err != nil {
		return []byte{}, err
	}

	// get certificate as bytes
	cert, err := decodeCert(data["certificate_authority_data"])
	if err != nil {
		return []byte{}, err
	}

	return []byte(cert), nil
}

func credentialsApi(clusterId string, accessToken string) (*http.Response, error) {
	httpClient := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf(digitalOceanApi, clusterId), nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := httpClient.Do(req)
	if err != nil {
		msg := fmt.Sprintf("failed to call digitalocean api: %s", err)
		return &http.Response{}, errors.New(msg)
	}
	if resp.StatusCode != http.StatusOK {
		msg := "non 2XX response from digitalocean api"
		return &http.Response{}, errors.New(msg)
	}
	return resp, nil
}

func parseCredentialsApiResponse(resp *http.Response) (map[string]string, error) {
	// parse response
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read digitalocean api response: %s", err)
		return make(map[string]string), errors.New(msg)
	}
	body := string(respBytes)
	var data map[string]string
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		msg := fmt.Sprintf("failed to unmarshal digitalocean api response json: %s", err)
		return make(map[string]string), errors.New(msg)
	}
	return data, nil
}

func decodeCert(cert string) ([]byte, error) {
	certBytes, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		msg := fmt.Sprintf("failed to decode cert: %s", err)
		return []byte{}, errors.New(msg)
	}
	return certBytes, nil
}
