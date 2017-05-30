namespace java com.uber.zanzibar.clients.baz

struct BazResponse {
    1: required string message
}

exception ServerErr {
    1: required string message
}
