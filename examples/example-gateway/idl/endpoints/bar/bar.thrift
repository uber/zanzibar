namespace java com.uber.zanzibar.clients.bar

include "../foo/foo.thrift"

typedef string UUID
typedef i64 (json.type = 'Date') Timestamp
typedef i64 (json.type = "Long") Long

enum Fruit {
    APPLE,
    BANANA
}

struct BarRequest {
    1: required string stringField
    2: required bool boolField
    3: required binary binaryField
    4: required Timestamp timestamp
    5: required Fruit enumField
    6: required Long longField
}

struct BarResponse {
    1: required string stringField (
        zanzibar.http.ref = "headers.some-header-field"
        zanzibar.validation.type = "object,number"
    )
    2: required i32 intWithRange
    3: required i32 intWithoutRange (zanzibar.ignore.integer.range = "true")
    4: required map<UUID, i32> mapIntWithRange
    5: required map<string, i32> mapIntWithoutRange (
        zanzibar.ignore.integer.range = "true"
    )
    6: required binary binaryField
}

struct QueryParamsStruct {
    1: required string name
    2: optional string userUUID
    3: optional string authUUID (zanzibar.http.ref="headers.x-uuid")
    // TODO: validate required behaviour for headers
    4: optional string authUUID2 (zanzibar.http.ref="headers.x-uuid2")
}

struct QueryParamsOptsStruct {
    1: required string name
    2: optional string userUUID
    3: optional string authUUID
    4: optional string authUUID2
}

struct ParamsStruct {
    1: required string userUUID (
        zanzibar.http.ref = "params.user-uuid"
        go.tag = "json:\"-\""
    )
}

exception BarException {
    1: required string stringField (zanzibar.http.ref = "headers.another-header-field")
}

service Bar {
    string helloWorld(
    ) throws (
       1: BarException barException (zanzibar.http.status = "403")
   ) (
       zanzibar.http.method = "GET"
       zanzibar.http.path = "/bar/hello"
       zanzibar.http.status = "200"
    )
    BarResponse normal (
        1: required BarRequest request
    ) throws (
        1: BarException barException (zanzibar.http.status = "403")
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/bar/bar-path"
        zanzibar.http.status = "200"
    )
    BarResponse noRequest (
    ) throws (
        1: BarException barException (zanzibar.http.status = "403")
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/bar/no-request-path"
        zanzibar.http.status = "200"
    )
    BarResponse missingArg (
    ) throws (
        1: BarException barException (zanzibar.http.status = "403")
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/bar/missing-arg-path"
        zanzibar.http.status = "200"
    )
    BarResponse tooManyArgs (
        1: required BarRequest request
        2: optional foo.FooStruct foo
    ) throws (
        1: BarException barException (zanzibar.http.status = "403")
        2: foo.FooException fooException (zanzibar.http.status = "418")
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/bar/too-many-args-path"
        zanzibar.http.status = "200"
        zanzibar.http.reqHeaders = "x-uuid,x-token"
        zanzibar.http.resHeaders = "x-uuid,x-token"
    )
    void argNotStruct (
        1: required string request
    ) throws (
        1: BarException barException (zanzibar.http.status = "403")
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/bar/arg-not-struct-path"
        zanzibar.http.status = "200"
        zanzibar.meta = "SomeMeta"
        zanzibar.handler = "bar.baz"
    )

    BarResponse argWithHeaders (
        1: required string name
        2: optional string userUUID (
            zanzibar.http.ref = "headers.x-uuid"
            go.tag = "json:\"-\""
        )
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.reqHeaders = "x-uuid"
        zanzibar.http.path = "/bar/argWithHeaders"
        zanzibar.http.status = "200"
    )

    BarResponse argWithQueryParams(
        1: required string name
        2: optional string userUUID
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/bar/argWithQueryParams"
        zanzibar.http.status = "200"
    )

    BarResponse argWithNestedQueryParams(
        1: required QueryParamsStruct request
        2: optional QueryParamsOptsStruct opt
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/bar/argWithNestedQueryParams"
        zanzibar.http.status = "200"
    )

    BarResponse argWithQueryHeader(
        1: optional string userUUID (
            zanzibar.http.ref = "headers.x-uuid"
        )
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/bar/argWithQueryHeader"
        zanzibar.http.status = "200"
    )

    // TODO: support params annotation
    BarResponse argWithParams(
        1: required string uuid (
            zanzibar.http.ref = "params.uuid"
            go.tag = "json:\"-\""
        )
        2: optional ParamsStruct params
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/bar/argWithParams/:uuid/segment/:user-uuid"
        zanzibar.http.status = "200"
    )

    BarResponse argWithManyQueryParams(
        1: required string aStr
        2: optional string anOptStr
        3: required bool aBool
        4: optional bool anOptBool
        5: required i8 aInt8
        6: optional i8 anOptInt8
        7: required i16 aInt16
        8: optional i16 anOptInt16
        9: required i32 aInt32
        10: optional i32 anOptInt32
        11: required i64 aInt64
        12: optional i64 anOptInt64
        13: required double aFloat64
        14: optional double anOptFloat64
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/bar/argWithManyQueryParams"
        zanzibar.http.status = "200"
    )
}
