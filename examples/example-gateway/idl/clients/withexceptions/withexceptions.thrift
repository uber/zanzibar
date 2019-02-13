namespace java com.uber.zanzibar.clients.withexceptions


exception ExceptionType1 {
    1: string message
}

service WithExceptionsService {
    string func1(
    ) throws (
        1: ExceptionType1 e1
    )
}
