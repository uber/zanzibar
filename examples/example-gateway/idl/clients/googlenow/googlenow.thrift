namespace java com.googlenow

// Service specification for generating Google Now client.
service GoogleNowService {
    void addCredentials(
        1: required string authCode
    ) (  // can throws exceptions here for non 2XX response
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/add-credentials"
        zanzibar.http.status = "202"
        zanzibar.http.reqHeaders = "x-uuid"
        zanzibar.http.resHeaders = "x-uuid"
    )
    void checkCredentials(
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/check-credentials"
        zanzibar.http.status = "202"
        // comma sparated list for required headers
        zanzibar.http.reqHeaders = "x-uuid"
        zanzibar.http.resHeaders = "x-uuid"
    )
}
