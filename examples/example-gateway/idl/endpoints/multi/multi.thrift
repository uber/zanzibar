namespace java com.uber.zanzibar.endpoints.multi

service ServiceAFront {
    string hello (
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/multi/serviceA_f/hello"
        zanzibar.http.status = "200"
    )
}

service ServiceBFront {
    string hello (
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/multi/serviceB_f/hello"
        zanzibar.http.status = "200"
    )
}

service ServiceCFront {
    string hello (
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/multi/serviceC_f/hello"
        zanzibar.http.status = "200"
    )
}