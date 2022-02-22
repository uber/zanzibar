namespace java com.uber.zanzibar.endpoints.authtoken

include "../models/meta.thrift"

struct AuthTokenResponse {
    1: required string access_token
    2: required i32 expires_in
    3: required string token_type
}

service AuthToken {
    AuthTokenResponse getAuthToken (
        1: required string authorization (
            zanzibar.http.ref = "headers.authorization"
            go.tag = "json:\"-\""
        )
        2: required string fastPlatformCredentials (
            zanzibar.http.ref = "headers.x-fast-platform-credentials"
            go.tag = "json:\"-\""
        )
    ) (
       zanzibar.http.method = "GET"
       zanzibar.http.path = "/fetch"
       zanzibar.http.status = "200"
       zanzibar.http.req.metadata = "meta.AuthorizationOnly,meta.FastPlatformCredentialsOnly"
    )
}

struct Product {
    1: required i32 id
    2: required string name
    3: required i32 year
    4: required string color
    5: required string pantone_value
}

service MultiCalls {
    Product getRandomProduct (
    ) (
       zanzibar.http.method = "GET"
       zanzibar.http.path = "/multi"
       zanzibar.http.status = "200"
    )
}