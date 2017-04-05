namespace java com.uber.zanzibar.clients.baz

struct BazRequest {
  1: required bool b1,
  2: required string s2,
  3: required i32 i3
}

struct BazResponse {
  1: required string message
}

exception SimpleErr {
  1: required string message
}

exception NewErr {
  1: required string message
}

service SimpleService {

  BazResponse Call(
    1: required BazRequest arg
  ) (
    zanzibar.http.status = "200"
    zanzibar.http.method = "POST"
    zanzibar.http.path = "/baz/call-path"
    zanzibar.handler = "baz.call"
  )

  void Simple() throws (
    1: SimpleErr simpleErr (zanzibar.http.status = "403")
  ) (
    zanzibar.http.status = "204"
    zanzibar.http.method = "GET"
    zanzibar.http.path = "/baz/simple-path"
    zanzibar.handler = "baz.simple"
  )

  void SimpleFuture() throws (
    1: SimpleErr simpleErr (zanzibar.http.status = "403")
    2: NewErr newErr (zanzibar.http.status = "404")
  ) (
    zanzibar.http.status = "204"
    zanzibar.http.method = "GET"
    zanzibar.http.path = "/baz/simple-future-path"
    zanzibar.handler = "baz.simpleFuture"
  )
}

// service SecondService {
//  string Echo(1: string arg)
// }
