namespace java com.uber.zanzibar.clients.baz

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
  BazResponse Compare(
    1: required BazRequest arg1
    2: required BazRequest arg2
  ) throws (
    1: AuthErr authErr
  )

  void Call(
    1: required BazRequest arg
  ) throws (
    1: AuthErr authErr
  )

  BazResponse Ping() ()

  void SillyNoop() throws (
    1: AuthErr authErr
    2: ServerErr serverErr
  )
}

// service SecondService {
//  string Echo(1: string arg)
// }
