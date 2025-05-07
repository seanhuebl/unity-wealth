package testmodels

type SignUpTest struct {
	Name               string
	ReqBody            string
	WantErrSubstr      string
	ExpectedStatusCode int
	ExpectedResponse   map[string]interface{}
}
