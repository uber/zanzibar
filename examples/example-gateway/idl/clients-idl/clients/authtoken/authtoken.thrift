namespace java com.uber.zanzibar.clients.authtoken

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
       zanzibar.http.method = "POST"
       zanzibar.http.path = "/dw/oauth2/access_token?client_id=58d9dd80-e9fa-46bb-ac53-9c41ed4f5499&grant_type=urn:demandware:params:oauth:grant-type:client-id:dwsid:dwsecuretoken"
       zanzibar.http.status = "200"
       zanzibar.http.reqHeaders = "authorization,x-fast-platform-credentials"
    )
}

struct Product {
    1: required i32 id
    2: required string name
    3: required i32 year
    4: required string color
    5: required string pantone_valu
}
typedef list<Product> ProductsList

struct ProductsResponse {
    1: required ProductsList data
}

struct ProductResponse {
    1: required Product data
}

service Products {
    ProductsResponse getProducts (
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/api/products"
        zanzibar.http.status = "304"
    )

    ProductResponse getProduct (
        1: required i32 product_id (zanzibar.http.ref = "params.product_id")
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/api/products/:product_id"
        zanzibar.http.status = "304"
    )
}