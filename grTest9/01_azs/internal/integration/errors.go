package integration

import (
	"fmt"
	"net/http"
)

type ErrKazsOperator struct {
	Url          string
	Request      *http.Request
	RequestBody  *[]byte
	Response     *http.Response
	ResponseBody *[]byte
	Message      string
}

func (o *ErrKazsOperator) Error() string {
	requestUrl := ""
	if o.Request != nil {
		requestUrl = fmt.Sprintf("%s %s", o.Request.Method, o.Request.URL)
	}

	statusCode := 0
	if o.Response != nil {
		statusCode = o.Response.StatusCode
	}

	responseBody := ""
	if o.ResponseBody != nil {
		responseBody = string(*o.ResponseBody)
	}

	return fmt.Sprintf("error: %s, requestUrl: %s, statusCode: %v, responceBody:%s", o.Message, requestUrl, statusCode, responseBody)
}

func NewErrKazsOperator(message, url string, request *http.Request, requestBody *[]byte, response *http.Response, responseBody *[]byte) *ErrKazsOperator {

	return &ErrKazsOperator{
		Url:          url,
		Request:      request,
		RequestBody:  requestBody,
		Response:     response,
		ResponseBody: responseBody,
		Message:      message,
	}
}

func ErrorUnwrap(err error) (*ErrKazsOperator, error) {
	unWrap, ok := err.(*ErrKazsOperator)
	if !ok {
		return nil, fmt.Errorf("errorKazsOperatorUnwrap - error")
	}
	return unWrap, nil
}
