package mantis

import (
	"testing"
	"github.com/skiptirengu/go-mantis-webhook/config"
	"github.com/stretchr/testify/assert"
)

type mockSoapRequester struct {
	Response string
}

func (m mockSoapRequester) request(method string, request string) ([]byte, error) {
	return []byte(m.Response), nil
}

func getRequesterMock(response string) (*mockSoapRequester) {
	return &mockSoapRequester{response}
}

func TestSoap_ProjectGetIdFromNameErr(t *testing.T) {
	service := NewSoapService(&config.Configuration{
		Mantis: config.MantisConfig{Host: "http://localhost"},
	})
	service.requester = getRequesterMock(`
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/">
   <SOAP-ENV:Body>
      <SOAP-ENV:Fault>
         <faultcode>SOAP-ENV:Client</faultcode>
         <faultstring>Access denied</faultstring>
      </SOAP-ENV:Fault>
   </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`)
	ret, err := service.ProjectGetIdFromName("test")
	assert.Equal(t, "Access denied", err.Error())
	assert.Equal(t, 0, ret)
}

func TestSoap_ProjectGetIdFromNameOk(t *testing.T) {
	service := NewSoapService(&config.Configuration{
		Mantis: config.MantisConfig{Host: "http://localhost"},
	})
	service.requester = getRequesterMock(`
<SOAP-ENV:Envelope SOAP-ENV:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ns1="http://futureware.biz/mantisconnect" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/">
   <SOAP-ENV:Body>
      <ns1:mc_project_get_id_from_nameResponse>
         <return xsi:type="xsd:integer">1</return>
      </ns1:mc_project_get_id_from_nameResponse>
   </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`)
	ret, err := service.ProjectGetIdFromName("test")
	assert.Nil(t, err)
	assert.Equal(t, 1, ret)
}

func TestSoap_ProjectGetUsersErr(t *testing.T) {
	service := NewSoapService(&config.Configuration{
		Mantis: config.MantisConfig{Host: "http://localhost"},
	})
	service.requester = getRequesterMock(`
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/">
   <SOAP-ENV:Body>
      <SOAP-ENV:Fault>
         <faultcode>SOAP-ENV:Client</faultcode>
         <faultstring>Hi chat :)</faultstring>
      </SOAP-ENV:Fault>
   </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`)
	res, err := service.ProjectGetUsers(0)
	assert.Equal(t, "Hi chat :)", err.Error())
	assert.Nil(t, res)
}

func TestSoap_ProjectGetUsersOk(t *testing.T) {
	service := NewSoapService(&config.Configuration{
		Mantis: config.MantisConfig{Host: "http://localhost"},
	})
	service.requester = getRequesterMock(`
<SOAP-ENV:Envelope SOAP-ENV:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ns1="http://futureware.biz/mantisconnect" xmlns:SOAP-ENC="http://schemas.xmlsoap.org/soap/encoding/" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
   <SOAP-ENV:Body>
      <ns1:mc_project_get_usersResponse>
         <return SOAP-ENC:arrayType="ns1:AccountData[2]" xsi:type="SOAP-ENC:Array">
            <item xsi:type="ns1:AccountData">
               <id xsi:type="xsd:integer">1</id>
               <name xsi:type="xsd:string">administrator</name>
               <email xsi:type="xsd:string">root@localhost</email>
            </item>
            <item xsi:type="ns1:AccountData">
               <id xsi:type="xsd:integer">2</id>
               <name xsi:type="xsd:string">Thiago</name>
               <real_name xsi:type="xsd:string">Thiago</real_name>
               <email xsi:type="xsd:string">thiago@example.com</email>
            </item>
         </return>
      </ns1:mc_project_get_usersResponse>
   </SOAP-ENV:Body>
</SOAP-ENV:Envelope>
`)
	res, err := service.ProjectGetUsers(1)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, AccountData{1, "administrator", "root@localhost"}, res[0])
	assert.Equal(t, AccountData{2, "Thiago", "thiago@example.com"}, res[1])
	assert.Nil(t, err)
}
