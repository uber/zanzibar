namespace java com.uber.zanzibar.clients.foo
include "base/base.thrift"

struct FooName {
    1: optional string name
}

struct FooStruct {
    1: required string fooString
    2: optional i32 fooI32
    3: optional i16 fooI16
    4: optional double fooDouble
    5: optional bool fooBool
    6: optional map<string, string> fooMap
    7: optional base.Message message
}

exception FooException {
    1: required string teapot
}