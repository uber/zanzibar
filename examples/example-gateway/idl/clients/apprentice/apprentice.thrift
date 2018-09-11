namespace java com.uber.zanzibar.clients.apprentice

exception OperationError {
    1: required string message
}

service Apprentice {
    string getRequest(
        1: required string traceID
        2: optional string clientServiceName
        3: optional string clientMethod
    ) throws (1: OperationError err)

    string getResponse(
        1: required string traceID
        2: optional string clientServiceName
        3: optional string clientMethod
    ) throws (1: OperationError err)
}
