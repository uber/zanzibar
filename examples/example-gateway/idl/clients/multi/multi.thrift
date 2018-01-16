namespace java com.uber.zanzibar.clients.multi

service ServiceABack {
    string hello (
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/multi/serviceA_b/hello"
        zanzibar.http.status = "200"
    )
}

service ServiceBBack {
    string hello (
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/multi/serviceB_b/hello"
        zanzibar.http.status = "200"
    )
}
