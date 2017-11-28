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

exception ServerErr {
    1: required string message
}
