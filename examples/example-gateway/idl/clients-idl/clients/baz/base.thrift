namespace java com.uber.zanzibar.clients.baz

typedef string UUID

struct NestedStruct {
    1: required string msg
    2: optional i32 check
}

struct BazResponse {
    1: required string message
}

struct TransStruct {
    1: required string message
    2: optional NestedStruct driver
    3: required NestedStruct rider
}

struct NestHeaders {
    1: required string UUID
    2: optional string token
}

struct Wrapped {
    1: required NestHeaders n1
    2: optional NestHeaders n2
}

struct TransHeaders {
    1: required Wrapped w1
    2: optional Wrapped w2
}

exception ServerErr {
    1: required string message
}
