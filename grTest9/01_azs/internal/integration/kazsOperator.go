package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type KazsOperator struct {
	requestUrl    url.URL
	password      string
	ConfigHash    string
	KazsID        string
	KazsNumber    string
	KazsTimezone  string
	NtpServer     string
	SupportNumber string
	Logo          string
}

func NewKazsOperator() (*KazsOperator, error) {

	api := &KazsOperator{}

	return api, nil
}

type sendRequestArgs struct {
	methodsKazsOperator MethodsKazsOperator
	urlValues           *url.Values
	requestBody         any
	resultStruct        any // Указатель на результирующую структуру
}

func (o *KazsOperator) setUrl(reqURL string) error {
	urlRes, err := url.ParseRequestURI("https://" + reqURL)
	if err != nil {
		return fmt.Errorf("url ParseRequestURI error: %v", err)
	}
	o.requestUrl = *urlRes

	return nil
}

func (o *KazsOperator) SetConfig(reqURL, pass, kazsID, kazsNumber, configHash string) error {
	urlRes, err := url.ParseRequestURI("https://" + reqURL)
	if err != nil {
		return fmt.Errorf("url ParseRequestURI error: %v", err)
	}
	o.requestUrl = *urlRes

	o.password = pass

	o.KazsID = kazsID

	o.KazsNumber = kazsNumber

	o.ConfigHash = configHash

	return nil
}

func (o *KazsOperator) SetGUI(kazsTimezone, ntpServer, supportNumber, logo string) error {
	o.KazsTimezone = kazsTimezone
	o.NtpServer = ntpServer
	o.SupportNumber = supportNumber
	o.Logo = logo

	return nil
}

func (o *KazsOperator) sendRequestKazsOperator(s sendRequestArgs) error {

	if o.requestUrl.String() == "" {
		return errors.New("Url is empty")
	}

	url := o.requestUrl
	url.Path = s.methodsKazsOperator.Path

	if s.urlValues != nil {
		url.RawQuery = s.urlValues.Encode()
	}
	requestUrl := url.String()

	var err error
	var requestBody []byte

	if s.requestBody != nil {
		requestBody, err = json.Marshal(s.requestBody)
		if err != nil {
			return NewErrKazsOperator(fmt.Sprintf("json_Marshal - error: %v", err), requestUrl, nil, nil, nil, nil)
		}
	}

	request, err := http.NewRequest(s.methodsKazsOperator.Method, requestUrl, strings.NewReader(string(requestBody)))
	if err != nil {
		return NewErrKazsOperator(fmt.Sprintf("http_NewRequest - error: %v", err), requestUrl, nil, &requestBody, nil, nil)
	}
	request.Header.Add("Authorization", o.password)
	request.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return NewErrKazsOperator(fmt.Sprintf("do - error: %v", err), requestUrl, request, &requestBody, nil, nil)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return NewErrKazsOperator(fmt.Sprintf("readAll - error: %v", err), requestUrl, request, &requestBody, response, &responseBody)
	}

	if response.StatusCode != 200 {
		return NewErrKazsOperator(fmt.Sprintf("statusCode: %s - error", response.Status), requestUrl, request, &requestBody, response, &responseBody)
	}

	if len(responseBody) > 0 {
		err = json.Unmarshal(responseBody, &s.resultStruct)
		if err != nil {
			return NewErrKazsOperator("json_Unmarshal - error", requestUrl, request, &requestBody, response, &responseBody)
		}
	}

	return nil
}
