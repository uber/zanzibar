namespace java com.uber.zanzibar.clients.baz
include "base.thrift"

typedef string UUID

enum Fruit {
   APPLE,
   BANANA
}

struct BazRequest {
    1: required bool b1
    2: required string s2
    3: required i32 i3
}

struct transHeaderType {
    1: required bool b1
    2: optional i32 i1
    3: required i64 i2
    4: optional double f3
    5: required UUID u4
    6: optional UUID u5
    7: required string s6
}
struct HeaderSchema {}

exception AuthErr {
    1: required string message
}

exception OtherAuthErr {
  1: required string message
}

struct Recur3 {
    1: required UUID field31
}

struct Recur2 {
    1: required Recur3 field21
    2: optional Recur3 field22
}

struct Recur1 {
    1: required map<UUID, Recur2> field1
}

struct Profile {
    1: required Recur1 recur1
}

struct GetProfileRequest {
    1: required UUID target
}

struct GetProfileResponse {
    1: required list<Profile> payloads
}

service SimpleService {
    GetProfileResponse getProfile(
        1: required GetProfileRequest request
    ) throws (
        1: AuthErr authErr
    )

    base.BazResponse compare(
        1: required BazRequest arg1
        2: required BazRequest arg2
    ) throws (
        1: AuthErr authErr
        2: OtherAuthErr otherAuthErr
    )

    base.TransStruct trans(
        1: required base.TransStruct arg1
        2: optional base.TransStruct arg2
    ) throws (
        1: AuthErr authErr
        2: OtherAuthErr otherAuthErr
    )

    base.TransHeaders transHeadersNoReq(
        1: required base.NestedStruct req
        2: optional string s2
        3: required i32 i3
        4: optional bool b4
    ) throws (
        1: AuthErr authErr
    )

    base.TransHeaders transHeaders(
        1: required base.TransHeaders req
    ) throws (
        1: AuthErr authErr
        2: OtherAuthErr otherAuthErr
    )

    transHeaderType transHeadersType(
        1: required transHeaderType req
    ) throws (
        1: AuthErr authErr
        2: OtherAuthErr otherAuthErr
    )

    HeaderSchema headerSchema(
        1: required HeaderSchema req
    ) throws (
        1: AuthErr authErr
        2: OtherAuthErr otherAuthErr
    )

    void call(
        1: required BazRequest arg
        2: optional i64 i64Optional (zanzibar.http.ref = "headers.x-token")
        3: optional UUID testUUID (zanzibar.http.ref = "headers.x-uuid")
    ) throws (
        1: AuthErr authErr
    ) (
        zanzibar.http.reqHeaders = "x-uuid,x-token"
        zanzibar.http.resHeaders = "some-res-header"
    )

     void anotherCall(
         1: required BazRequest arg
         2: optional i64 i64Optional (zanzibar.http.ref = "headers.x-token")
         3: optional UUID testUUID (zanzibar.http.ref = "headers.x-uuid")
     ) throws (
         1: AuthErr authErr
     ) (
         zanzibar.http.reqHeaders = "x-uuid,x-token"
         zanzibar.http.resHeaders = "some-res-header"
     )

    base.BazResponse ping() ()

    void sillyNoop() throws (
        1: AuthErr authErr
        2: base.ServerErr serverErr
    )

    void testUuid()
    void urlTest()
}

service SecondService {
    i8 echoI8 (
        1: required i8 arg
    )

    i16 echoI16(
        1: required i16 arg
    )

    i32 echoI32(
        1: required i32 arg
    )

    i64 echoI64(
        1: required i64 arg
    )

    double echoDouble(
        1: required double arg
    )

    bool echoBool (
        1: required bool arg
    )

    binary echoBinary (
        1: required binary arg
    )

    string echoString(
        1: required string arg
    )

    Fruit echoEnum (
        1: required Fruit arg = Fruit.APPLE
    )

    base.UUID echoTypedef(
        1: required base.UUID arg
    )

    set<string> echoStringSet(
        1: required set<string> arg
    )

    // value is unhashable
    set<base.BazResponse> echoStructSet(
        1: required set<base.BazResponse> arg
    )

    list<string> echoStringList (
        1: required list<string> arg
    )

    list<base.BazResponse> echoStructList (
        1: required list<base.BazResponse> arg
    )

    map<string, base.BazResponse> echoStringMap (
        1: required map<string, base.BazResponse> arg
    )

    // key is unhashable
    map<base.BazResponse, string> echoStructMap (
        1: required map<base.BazResponse, string> arg
    )
}
