namespace java com.uber.zanzibar.clients.baz

typedef string UUID

struct BazRequest {
  1: required bool b1
  2: required string s2
  3: required i32 i3
}

struct BazResponse {
  1: required string message
}

exception AuthErr {
  1: required string message
}

exception ServerErr {
  1: required string message
}

service SimpleService {
  // have both request body and response body
  BazResponse Compare(
    1: required BazRequest arg1
    2: required BazRequest arg2
  ) throws (
    1: AuthErr authErr (zanzibar.http.status = "403")
  ) (
    zanzibar.http.status = "200"
    zanzibar.http.method = "POST"
    zanzibar.http.path = "/baz/compare"
    zanzibar.handler = "baz.compare"
  )

  string Echo(
    1: required string msg
  )

  // no response body
  void Call(
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

  // no response body
  void AnotherCall(
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
  BazResponse Ping() (
    zanzibar.http.status = "200"
    zanzibar.http.method = "GET"
    zanzibar.http.path = "/baz/ping"
    zanzibar.handler = "baz.multiArgs"
  )

  // neither request body nor response body
  void SillyNoop() throws (
    1: AuthErr authErr (zanzibar.http.status = "403")
    2: ServerErr serverErr (zanzibar.http.status = "500")
  ) (
    zanzibar.http.status = "204"
    zanzibar.http.method = "GET"
    zanzibar.http.path = "/baz/silly-noop"
    zanzibar.handler = "baz.sillyNoop"
  )
}

// service SecondService {
//  string Echo(1: string arg)
// }
