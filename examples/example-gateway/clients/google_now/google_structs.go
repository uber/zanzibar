package googleNow

//go:generate easyjson -all $GOFILE

// AddCredentialRequest is the request for AddCredential.
type AddCredentialRequest struct {
	AuthCode string `json:"authCode"`
}

// AddCredentialResponse is the response for AddCredential.
type AddCredentialResponse struct {
	StatusCode int `json:"statusCode"`
}
