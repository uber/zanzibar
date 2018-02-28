namespace java com.uber.zanzibar.clients.baz

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

service SimpleService {
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
    zanzibar.http.reqHeaders = "x-uuid,x-token"
    zanzibar.http.resHeaders = "some-res-header"
  )

  // no request body
  BazResponse ping() (
    zanzibar.http.status = "200"
    zanzibar.http.method = "GET"
    zanzibar.http.path = "/baz/ping"
    zanzibar.handler = "baz.multiArgs"
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
