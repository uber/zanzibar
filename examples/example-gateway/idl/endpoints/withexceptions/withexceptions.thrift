namespace java com.uber.zanzibar.clients.withexceptions


exception EndpointExceptionType1 {
    1: required string message1
}

exception EndpointExceptionType2 {
    1: required string message2
}

struct Response {}

service WithExceptions {
    Response Func1(
    ) throws (
        1: EndpointExceptionType1 e1 (zanzibar.http.status = "401")
        2: EndpointExceptionType2 e2 (zanzibar.http.status = "401")
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/withexceptions/func1"
        zanzibar.http.status = "200"
    )
}
