namespace java com.uber.zanzibar.clients.bar

include "../foo/foo.thrift"

struct BarRequest {
    1: required string stringField (
        zanzibar.http.ref = "params.someParamsField"
    )
    2: required bool boolField (zanzibar.http.ref = "query.some-query-field")
}
struct BarResponse {
    1: required string stringField (
        zanzibar.http.ref = "headers.some-header-field"
        zanzibar.validation.type = "object,number"
    )
    2: required i32 intWithRange
    3: required i32 intWithoutRange (zanzibar.ignore.integer.range = "true")
    4: required map<string, i32> mapIntWithRange
    5: required map<string, i32> mapIntWithoutRange (
        zanzibar.ignore.integer.range = "true"
    )
}

exception BarException {
    1: required string stringField (
        zanzibar.http.ref = "headers.another-header-field"
    )
}

service Bar {
    BarResponse normal (
        1: required BarRequest request
    ) throws (
        1: BarException barException (zanzibar.http.status = "403")
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/bar-path"
        zanzibar.http.status = "200"
    )
    BarResponse noRequest (
    ) throws (
        1: BarException barException (zanzibar.http.status = "403")
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/no-request-path"
        zanzibar.http.status = "200"
    )
    BarResponse missingArg (
    ) throws (
        1: BarException barException (zanzibar.http.status = "403")
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/missing-arg-path"
        zanzibar.http.status = "200"
    )
    BarResponse tooManyArgs (
        1: required BarRequest request
        2: optional foo.FooStruct foo
    ) throws (
        1: BarException barException (zanzibar.http.status = "403")
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/too-many-args-path"
        zanzibar.http.status = "200"
    )
    void argNotStruct (
        1: required string request
    ) throws (
        1: BarException barException (zanzibar.http.status = "403")
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/arg-not-struct-path"
        zanzibar.http.status = "200"
    )

    BarResponse argWithHeaders (
        1: required string name
        2: optional string userUUID
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.reqHeaders = "x-uuid"
        zanzibar.http.path = "/bar/argWithHeaders"
        zanzibar.http.status = "200"
    )

    string echo (
        1: required string name
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.reqHeaders = "x-uuid"
        zanzibar.http.path = "/bar/echo"
        zanzibar.http.status = "200"
    )
}
