package mantis

import (
	"fmt"
	"bytes"
	"encoding/xml"
	"errors"
	"github.com/parnurzeal/gorequest"
	"github.com/skiptirengu/go-mantis-webhook/config"
	"html/template"
	"github.com/skiptirengu/go-mantis-webhook/util"
	"io/ioutil"
)

const (
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

	mcProjectGetIdFromNameXML = `<soapenv:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:man="http://futureware.biz/mantisconnect">
		<soapenv:Header/>
		<soapenv:Body>
			<man:mc_project_get_id_from_name soapenv:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
				<username xsi:type="xsd:string">{{.Username}}</username>
				<password xsi:type="xsd:string">{{.Password}}</password>
				<project_name xsi:type="xsd:string">{{.ProjectName}}</project_name>
			</man:mc_project_get_id_from_name>
		</soapenv:Body>
	</soapenv:Envelope>`
)

type AccountData struct {
	Id    int    `xml:"id"`
	Name  string `xml:"name"`
	Email string `xml:"email"`
}

type ProjectGetUsersResponse struct {
	Accounts []AccountData `xml:"Body>mc_project_get_usersResponse>return>item"`
}

type ProjectGetIdFromNameResponse struct {
	ID int `xml:"Body>mc_project_get_id_from_nameResponse>return"`
}

type FaultResponse struct {
	FaultCode   string `xml:"Body>Fault>faultcode"`
	FaultString string `xml:"Body>Fault>faultstring"`
}

type projectGetUsersRequest struct {
	Username  string
	Password  string
	ProjectID int
}

type projectGetIdFromNameRequest struct {
	Username    string
	Password    string
	ProjectName string
}

type soap struct {
	conf      *config.Configuration
	requester soapRequester
}

type soapRequester interface {
	request(method string, request string) ([]byte, error)
}

type defaultSoapRequester struct{}

func (defaultSoapRequester) request(method string, request string) ([]byte, error) {
	resp, _, errs := gorequest.New().Post(method).Type("xml").
		AppendHeader("Content-Type", "text/xml;charset=UTF-8").
		SendString(request).End()

	if err := util.PopError(errs); err != nil {
		return nil, err
	}

	return ioutil.ReadAll(resp.Body)
}

func NewSoapService(c *config.Configuration) (*soap) {
	return &soap{c, &defaultSoapRequester{}}
}

func (s soap) soapEndpoint() (string) {
	return fmt.Sprintf("%s/api/soap/mantisconnect.php", getHost(s.conf))
}

func (s soap) soapAction(method string) (string) {
	return fmt.Sprintf("%s/%s", s.soapEndpoint(), method)
}

func (s soap) ProjectGetIdFromName(name string) (int, error) {
	var (
		params = projectGetIdFromNameRequest{s.conf.Mantis.User, s.conf.Mantis.Password, name}
		resp   = &ProjectGetIdFromNameResponse{}
	)

	if err := s.xmlMakeSoapRequest("mc_project_get_id_from_name", mcProjectGetIdFromNameXML, params, resp); err != nil {
		return 0, err
	} else {
		return resp.ID, nil
	}
}

func (s soap) ProjectGetUsers(projectId int) ([]AccountData, error) {
	var (
		params = projectGetUsersRequest{s.conf.Mantis.User, s.conf.Mantis.Password, projectId}
		resp   = &ProjectGetUsersResponse{}
	)

	if err := s.xmlMakeSoapRequest("mc_project_get_users", mcProjectGetUsersXML, params, resp); err != nil {
		return nil, err
	} else {
		return resp.Accounts, nil
	}
}

func (s soap) xmlMakeSoapRequest(method, xmlTemplate string, params, result interface{}) (error) {
	var body = bytes.NewBufferString("")

	requestXML := template.Must(template.New(method).Parse(xmlTemplate))

	if err := requestXML.Execute(body, params); err != nil {
		return err
	}

	data, err := s.makeRequest("mc_project_get_users", string(body.Bytes()))
	if err != nil {
		return err
	}

	return xml.Unmarshal(data, result)
}

func (s soap) makeRequest(method string, request string) ([]byte, error) {
	var (
		data  []byte
		fault = &FaultResponse{}
	)

	data, err := s.requester.request(s.soapAction(method), request)
	if err != nil {
		return nil, err
	}

	xml.Unmarshal(data, fault)
	if fault.FaultCode != "" {
		return nil, errors.New(fault.FaultString)
	}

	return data, nil
}
