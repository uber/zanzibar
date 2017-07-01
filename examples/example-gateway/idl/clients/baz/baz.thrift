namespace java com.uber.zanzibar.clients.baz
include "base.thrift"

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
    string Echo(
        1: required string arg
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
