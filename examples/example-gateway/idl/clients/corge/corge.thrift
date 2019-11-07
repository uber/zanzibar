namespace java com.uber.zanzibar.clients.corge

exception NotModified { }
struct Foo { }

service Corge {
    string echoString(
        1: required string arg
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/echo/string"
        zanzibar.http.status = "200"
    )
    // this method is intentionally not exposed in client-config.yaml
    bool echoBool(
        1: required bool arg
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/echo/bool"
        zanzibar.http.status = "200"
    )
    void noContent(
        1: required bool arg
    ) throws (
        1: NotModified notModified (zanzibar.http.status = "304")
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/echo/no-content"
        zanzibar.http.status = "204"
    )
    void noContentNoException(
        1: required bool arg
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/echo/no-content-no-exception"
        zanzibar.http.status = "204"
    )
    Foo noContentOnException(
        1: required bool arg
    ) throws (
        1: NotModified notModified (zanzibar.http.status = "304")
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/echo/no-content-on-exception"
        zanzibar.http.status = "200"
    )
}
