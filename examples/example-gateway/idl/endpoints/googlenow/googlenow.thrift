namespace java com.uber.zanzibar.clients.googlenow

include "../models/meta.thrift"

// Service specification for generating Google Now client.
service GoogleNow {
    void addCredentials(
        1: required string authCode
    ) (  // can throws exceptions here for non 2XX response
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/googlenow/add-credentials"
        zanzibar.http.status = "202"
        zanzibar.http.req.metadata = "meta.UUIDOnly,meta.TokenOnly"
        zanzibar.http.res.metadata = "meta.UUIDOnly"
    )
    void checkCredentials(
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/googlenow/check-credentials"
        zanzibar.http.status = "202"
        zanzibar.http.req.metadata = "meta.UUIDOnly,meta.TokenOnly"
        zanzibar.http.res.metadata = "meta.UUIDOnly"
    )
}
