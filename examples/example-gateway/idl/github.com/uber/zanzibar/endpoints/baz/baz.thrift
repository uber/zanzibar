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
  BazResponse Call(1: required BazRequest arg)
  void Simple() throws (1: SimpleErr simpleErr)
  void SimpleFuture() throws (1: SimpleErr simpleErr, 2: NewErr newErr)
}

// service SecondService {
//  string Echo(1: string arg)
// }
