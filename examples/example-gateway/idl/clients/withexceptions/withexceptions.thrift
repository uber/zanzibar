namespace java com.uber.zanzibar.clients.withexceptions


exception ExceptionType1 {
    1: required string message
}

service WithExceptionsService {
    string func1(
    ) throws (
        1: ExceptionType1 e1 (zanzibar.http.status = "401")
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/withexceptions/func1"
        zanzibar.http.status = "200"
    )
}
