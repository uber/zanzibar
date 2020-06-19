namespace java com.uber.zanzibar.clients.baz

include "../models/meta.thrift"

typedef string UUID

struct BazRequest {
  1: required bool b1
  2: required string s2
  3: required i32 i3
}

struct NestedStruct {
    1: required string msg
    2: optional i32 check
}

struct TransStruct {
    1: required string message
    2: optional NestedStruct driver
    3: required NestedStruct rider
}

struct TransHeader {}

struct HeaderSchema {}

struct BazResponse {
  1: required string message
}

exception AuthErr {
  1: required string message
}

exception OtherAuthErr {
  1: required string message
}

exception ServerErr {
  1: required string message
}

struct GetProfileRequest {
    1: required UUID target
}

struct GetProfileResponse {
    1: required list<Profile> payloads
}

struct Profile {
    1: required Recur1 recur1
}

struct Recur1 {
    1: required map<UUID, Recur2> field1
}

struct Recur2 {
    1: required Recur3 field21
    2: required Recur3 field22
}

struct Recur3 {
    1: required UUID field31
}

service SimpleService {
  GetProfileResponse getProfile(
    1: required GetProfileRequest request
  ) throws (
    1: AuthErr authErr (zanzibar.http.status = "403")
  ) (
    zanzibar.http.status = "200"
    zanzibar.http.method = "POST"
    zanzibar.http.path = "/baz/get-profile"
  )
  // have both request body and response body
  BazResponse compare(
    1: required BazRequest arg1
    2: required BazRequest arg2
  ) throws (
    1: AuthErr authErr (zanzibar.http.status = "403")
    2: OtherAuthErr otherAuthErr (zanzibar.http.status = "403")
  ) (
    zanzibar.http.status = "200"
    zanzibar.http.method = "POST"
    zanzibar.http.path = "/baz/compare"
    zanzibar.handler = "baz.compare"
  )

  TransStruct trans(
      1: required TransStruct arg1
      2: optional TransStruct arg2
  ) throws (
      1: AuthErr authErr (zanzibar.http.status = "403")
      2: OtherAuthErr otherAuthErr (zanzibar.http.status = "403")
  ) (
      zanzibar.http.status = "200"
      zanzibar.http.method = "POST"
      zanzibar.http.path = "/baz/trans"
      zanzibar.handler = "baz.trans"
  )

  TransHeader transHeadersNoReq() throws (
      1: AuthErr authErr (zanzibar.http.status = "401")
  ) (
      zanzibar.http.status = "200"
      zanzibar.http.method = "POST"
      zanzibar.http.path = "/baz/trans-headers-no-req"
      zanzibar.http.req.metadata = "meta.Dgx"
  )

  TransHeader transHeaders(
      1: required TransHeader req
  ) throws (
      1: AuthErr authErr (zanzibar.http.status = "401")
      2: OtherAuthErr otherAuthErr (zanzibar.http.status = "403")
  ) (
      zanzibar.http.status = "200"
      zanzibar.http.method = "POST"
      zanzibar.http.path = "/baz/trans-headers"
      zanzibar.http.req.metadata = "meta.Garply"
  )

  HeaderSchema headerSchema(
      1: required HeaderSchema req
  ) throws (
      1: AuthErr authErr (zanzibar.http.status = "401")
      2: OtherAuthErr otherAuthErr (zanzibar.http.status = "403")
  ) (
      zanzibar.http.status = "200"
      zanzibar.http.method = "POST"
      zanzibar.http.path = "/baz/header-schema"
      zanzibar.http.req.metadata = "meta.Grault,meta.Fred"
  )


  TransHeader transHeadersType(
      1: required TransHeader req
  ) throws (
      1: AuthErr authErr (zanzibar.http.status = "401")
      2: OtherAuthErr otherAuthErr (zanzibar.http.status = "403")
  ) (
      zanzibar.http.status = "200"
      zanzibar.http.method = "POST"
      zanzibar.http.path = "/baz/trans-header-type"
      zanzibar.http.reqHeaders = "x-boolean,x-int,x-float,x-string"
  )

  // no response body
  void call(
    1: required BazRequest arg
    2: optional i64 i64Optional (zanzibar.http.ref = "headers.x-token")
    3: optional UUID testUUID (zanzibar.http.ref = "headers.x-uuid")
  ) throws (
    1: AuthErr authErr (zanzibar.http.status = "403")
  ) (
    zanzibar.http.status = "204"
    zanzibar.http.method = "POST"
    zanzibar.http.path = "/baz/call"
    zanzibar.handler = "baz.call"
    zanzibar.http.req.metadata = "meta.Grault"
    zanzibar.http.res.metadata = "meta.Thud"
  )

 // no response body
  void anotherCall(
    1: required BazRequest arg
    2: optional i64 i64Optional (zanzibar.http.ref = "headers.x-token")
    3: optional UUID testUUID (zanzibar.http.ref = "headers.x-uuid")
  ) throws (
    1: AuthErr authErr (zanzibar.http.status = "403")
  ) (
    zanzibar.http.status = "204"
    zanzibar.http.method = "POST"
    zanzibar.http.path = "/baz/call"
    zanzibar.handler = "baz.call"
    zanzibar.http.req.metadata = "meta.Grault"
    zanzibar.http.res.metadata = "meta.Thud"
  )

  // no request body
  BazResponse ping() (
    zanzibar.http.status = "200"
    zanzibar.http.method = "GET"
    zanzibar.http.path = "/baz/ping"
    zanzibar.handler = "baz.multiArgs"
  )

  // neither request body nor response body
  void sillyNoop() throws (
    1: AuthErr authErr (zanzibar.http.status = "403")
    2: ServerErr serverErr (zanzibar.http.status = "500")
  ) (
    zanzibar.http.status = "204"
    zanzibar.http.method = "GET"
    zanzibar.http.path = "/baz/silly-noop"
    zanzibar.handler = "baz.sillyNoop"
  )

  void testUuid() (
    zanzibar.http.status = "204"
    zanzibar.http.method = "GET"
    zanzibar.http.path = "/baz/test-uuid"
    zanzibar.handler = "baz.testUuid"
  )

  void urlTest() (
    zanzibar.http.status = "204"
    zanzibar.http.method = "GET"
    zanzibar.http.path = "/baz/url-test"
    zanzibar.handler = "baz.urlTest"
  )
}
