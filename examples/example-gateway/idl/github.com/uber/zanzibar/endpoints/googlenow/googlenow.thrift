namespace java com.uber.zanzibar.clients.googlenow

// Service specification for generating Google Now client.
service GoogleNow {
    void addCredentials(
        1: required string authCode
    ) (  // can throws exceptions here for non 2XX response
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/googlenow/add-credentials"
        zanzibar.http.status = "202"
        zanzibar.http.headers = "x-uuid,x-token"
    )
    void checkCredentials(
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/googlenow/check-credentials"
        zanzibar.http.status = "202"
        // comma sparated list for required headers
        zanzibar.http.headers = "x-uuid,x-token"
    )
}
