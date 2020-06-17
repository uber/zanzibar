namespace java com.uber.zanzibar.endpoints.tchannel.quux

service SimpleService {
  string EchoString(
    1: required string msg
  )
}
