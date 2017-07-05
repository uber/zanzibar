namespace java com.uber.zanzibar.clients.baz
include "base.thrift"

enum Fruit {
   APPLE,
   BANANA,
   PEACH,
   GRAPE
}

struct BazRequest {
    1: required bool b1
    2: required string s2
    3: required i32 i3
}

exception AuthErr {
    1: required string message
}

exception OtherAuthErr {
  1: required string message
}

service SimpleService {
    base.BazResponse Compare(
        1: required BazRequest arg1
        2: required BazRequest arg2
    ) throws (
        1: AuthErr authErr
        2: OtherAuthErr otherAuthErr
    )

    void Call(
        1: required BazRequest arg
    ) throws (
        1: AuthErr authErr
    ) (
        zanzibar.http.reqHeaders = "x-uuid,x-token"
        zanzibar.http.resHeaders = "some-res-header"
    )

    base.BazResponse Ping() ()

    void SillyNoop() throws (
        1: AuthErr authErr
        2: base.ServerErr serverErr
    )
}

service SecondService {
    i8 EchoI8(
        1: required i8 arg
    )

    i16 EchoI16(
        1: required i16 arg
    )

    i32 EchoI32(
        1: required i32 arg
    )

    i64 EchoI64(
        1: required i64 arg
    )

    double EchoDouble(
        1: required double arg
    )

    bool EchoBool (
        1: required bool arg
    )

    binary EchoBinary (
        1: required binary arg
    )

    string EchoString(
        1: required string arg
    )

    Fruit EchoEnum (
        1: required Fruit arg = Fruit.APPLE
    )

    base.UUID EchoUUID(
        1: required base.UUID arg
    )

    list<string> EchoList(
        1: required list<string> arg
    )

    list<base.UUID> EchoUUIDList(
        1: required list<base.UUID> arg
    )

    set<string> EchoSet(
        1: required set<string> arg
    )

    map<base.UUID, base.BazResponse> EchoMap(
        1: required map<base.UUID, string> arg
    )
}
