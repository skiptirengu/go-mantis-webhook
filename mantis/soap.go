package mantis

import (
	"fmt"
	"bytes"
	"encoding/xml"
	"html/template"
	"github.com/skiptirengu/go-mantis-webhook/config"
	"errors"
	"github.com/parnurzeal/gorequest"
)

var (
	mcProjectGetUsersXML = `<soapenv:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:man="http://futureware.biz/mantisconnect">
   		<soapenv:Header/>
   		<soapenv:Body>
      		<man:mc_project_get_users soapenv:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
				<username xsi:type="xsd:string">{{.Username}}</username>
				<password xsi:type="xsd:string">{{.Password}}</password>
         		<project_id xsi:type="xsd:integer">{{.ProjectID}}</project_id>
         		<access xsi:type="xsd:integer">0</access>
      		</man:mc_project_get_users>
   		</soapenv:Body>
	</soapenv:Envelope>`
)

func soapEndpoint() (string) {
	return fmt.Sprintf("%s/api/soap/mantisconnect.php", getHost())
}

func soapAction(method string) (string) {
	return fmt.Sprintf("%s/%s", soapEndpoint(), method)
}

func ProjectGetUsers(projectId int) (*ProjectGetUsersResponse, error) {
	var (
		mantisConfig  = config.Get().Mantis
		requestBody   = bytes.NewBufferString("")
		requestParams = projectGetUsersRequest{
			Username:  mantisConfig.User,
			Password:  mantisConfig.Password,
			ProjectID: projectId,
		}
	)

	requestXML := template.Must(template.New("ProjectGetUsers").Parse(mcProjectGetUsersXML))
	if err := requestXML.Execute(requestBody, requestParams); err != nil {
		return nil, err
	}

	data, err := SoapRequest("mc_project_get_users", string(requestBody.Bytes()))
	if err != nil {
		return nil, err
	}

	resp := &ProjectGetUsersResponse{}
	err = xml.Unmarshal(data, resp)

	return resp, err
}

func SoapRequest(method string, request string) ([]byte, error) {
	var (
		data  []byte
		fault = &faultResponse{}
	)

	_, body, errs := gorequest.New().Post(soapAction(method)).Type("xml").
		AppendHeader("Content-Type", "text/xml;charset=UTF-8").
		SendString(request).End()

	if errLen := len(errs); errLen > 0 {
		return nil, errs[errLen-1]
	}

	data = bytes.NewBufferString(body).Bytes()
	xml.Unmarshal(data, fault)

	if fault.FaultCode != "" {
		return nil, errors.New(fault.FaultString)
	}

	return data, nil
}

type AccountData struct {
	Id    int    `xml:"id"`
	Name  string `xml:"name"`
	Email string `xml:"email"`
}

type ProjectGetUsersResponse struct {
	Result []AccountData `xml:"Body>mc_project_get_usersResponse>return>item"`
}

type faultResponse struct {
	FaultCode   string `xml:"Body>Fault>faultcode"`
	FaultString string `xml:"Body>Fault>faultstring"`
}

type projectGetUsersRequest struct {
	Username  string
	Password  string
	ProjectID int
}
