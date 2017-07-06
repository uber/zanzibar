namespace java com.uber.zanzibar.clients.baz

typedef string UUID

struct BazResponse {
    1: required string message
}

exception ServerErr {
    1: required string message
}
