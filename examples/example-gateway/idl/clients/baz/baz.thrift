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

service SimpleService {
  base.BazResponse Compare(
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

  base.BazResponse Ping() ()

  void SillyNoop() throws (
    1: AuthErr authErr
    2: base.ServerErr serverErr
  )
}

service SecondService {
  void Echo(
    1: required string arg
  )
}