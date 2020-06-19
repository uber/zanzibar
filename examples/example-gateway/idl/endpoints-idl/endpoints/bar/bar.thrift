namespace java com.uber.zanzibar.clients.bar

include "../foo/foo.thrift"
include "../models/meta.thrift"

typedef string UUID
typedef i64 (json.type = 'Date') Timestamp
typedef i64 (json.type = "Long") Long
typedef list<string> StringList
typedef list<UUID> UUIDList

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

struct RequestWithDuplicateType {
    1: optional BarRequest request1
    2: optional BarRequest request2
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
    7: optional BarResponse nextResponse
}

struct QueryParamsStruct {
    1: required string name
    2: optional string userUUID
    3: optional string authUUID (zanzibar.http.ref="headers.x-uuid")
    4: optional string authUUID2 (zanzibar.http.ref="headers.x-uuid2")
    5: required list<string> foo
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

enum DemoType {
    FIRST,
    SECOND
}

exception BarException {
    1: required string stringField (zanzibar.http.ref = "headers.another-header-field")
}

exception SeeOthersRedirection {
}

service Bar {
    string helloWorld(
    ) throws (
       1: BarException barException (zanzibar.http.status = "403")
       2: SeeOthersRedirection seeOthersRedirection (zanzibar.http.status = "303", zanzibar.http.res.body.disallow = "true")
   ) (
       zanzibar.http.method = "GET"
       zanzibar.http.path = "/bar/hello"
       zanzibar.http.status = "200"
    )

    string listAndEnum (
        1: required list<string> demoIds
        2: optional DemoType demoType
        3: optional list<DemoType> demos
    ) throws (
        1: BarException barException (zanzibar.http.status = "403")
    ) (
       zanzibar.http.method = "GET"
       zanzibar.http.path = "/bar/list-and-enum"
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
        zanzibar.http.req.metadata = "meta.Grault"
        zanzibar.http.res.metadata = "meta.Grault"
    )
    void argNotStruct (
        1: required string request
    ) throws (
        1: BarException barException (zanzibar.http.status = "403")
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/bar/arg-not-struct-path"
        zanzibar.http.status = "200"
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
        zanzibar.http.path = "/bar/argWithHeaders"
        zanzibar.http.req.metadata = "meta.UUIDOnly"
        zanzibar.http.status = "200"
    )

    BarResponse argWithQueryParams(
        1: required string name
        2: optional string userUUID
        3: optional list<string> foo
        4: required list<i8> bar
        // TODO: actually use sets here
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/bar/argWithQueryParams"
        zanzibar.http.status = "200"
        zanzibar.http.req.metadata = "meta.Grault"
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
        1: required string aStr		(zanzibar.http.ref = "query.aStr")
        2: optional string anOptStr	(zanzibar.http.ref = "query.anOptStr")
        3: required bool aBool		(zanzibar.http.ref = "query.aBool")
        4: optional bool anOptBool	(zanzibar.http.ref = "query.anOptBool")
        5: required i8 aInt8		(zanzibar.http.ref = "query.aInt8")
        6: optional i8 anOptInt8	(zanzibar.http.ref = "query.anOptInt8")
        7: required i16 aInt16		(zanzibar.http.ref = "query.aInt16")
        8: optional i16 anOptInt16	(zanzibar.http.ref = "query.anOptInt16")
        9: required i32 aInt32		(zanzibar.http.ref = "query.aInt32")
        10: optional i32 anOptInt32	(zanzibar.http.ref = "query.anOptInt32")
        11: required i64 aInt64		(zanzibar.http.ref = "query.aInt64")
        12: optional i64 anOptInt64	(zanzibar.http.ref = "query.anOptInt64")
        13: required double aFloat64	(zanzibar.http.ref = "query.aFloat64")
        14: optional double anOptFloat64	(zanzibar.http.ref = "query.anOptFloat64")
        15: required UUID aUUID		(zanzibar.http.ref = "query.aUUID")
        16: optional UUID anOptUUID	(zanzibar.http.ref = "query.anOptUUID")
        17: required list<UUID> aListUUID	(zanzibar.http.ref = "query.aListUUID")
        18: optional list<UUID> anOptListUUID	(zanzibar.http.ref = "query.anOptListUUID")
        19: required StringList aStringList	(zanzibar.http.ref = "query.aStringList")
        20: optional StringList anOptStringList	(zanzibar.http.ref = "query.anOptStringList")
        21: required UUIDList aUUIDList	(zanzibar.http.ref = "query.aUUIDList")
        22: optional UUIDList anOptUUIDList	(zanzibar.http.ref = "query.anOptUUIDList")
        23: required Timestamp aTs	(zanzibar.http.ref = "query.aTs")
        24: optional Timestamp anOptTs	(zanzibar.http.ref = "query.anOptTs")
        25: required DemoType aReqDemo
        26: optional Fruit anOptFruit
        27: required list<Fruit> aReqFruits
        28: optional list<DemoType> anOptDemos
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/bar/argWithManyQueryParams"
        zanzibar.http.status = "200"
    )

    BarResponse argWithParamsAndDuplicateFields(
        1: required RequestWithDuplicateType request
        2: required string entityUUID (zanzibar.http.ref = "params.uuid")
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/bar/argWithParamsAndDuplicateFields/:uuid/segment"
        zanzibar.http.status = "200"
    )

    BarResponse argWithNearDupQueryParams(
        1: required string one   (zanzibar.http.ref = "query.oneName")
        2: optional i32 two      (zanzibar.http.ref = "query.one_name")
        3: string three          (zanzibar.http.ref = "query.One_NamE")
        4: optional string four  (zanzibar.http.ref = "query.one-Name")
    ) (
        zanzibar.http.method = "GET"
        zanzibar.http.path = "/bar/argWithNearDupQueryParams"
        zanzibar.http.status = "200"
    )

    void deleteWithQueryParams(
      2: required string filter (zanzibar.http.ref = "query.filter")
      3: optional i32 count (zanzibar.http.ref = "query.count")
    ) (
       zanzibar.http.method = "DELETE"
       zanzibar.http.path = "/bar/withQueryParams"
       zanzibar.http.status = "200"
    )

    void deleteWithBody(
      1: required string filter
      2: optional i32 count
    ) (
      zanzibar.http.method = "DELETE"
      zanzibar.http.path = "/bar/withBody"
      zanzibar.http.status = "200"
    )
}
