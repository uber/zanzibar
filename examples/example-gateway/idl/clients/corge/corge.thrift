namespace java com.uber.zanzibar.clients.corge

service Corge {
    string echoString(
        1: required string arg
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/echo/string"
        zanzibar.http.status = "200"
    )
    // this method is intentionally not exposed in client-config.json
    bool echoBool(
        1: required bool arg
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/echo/bool"
        zanzibar.http.status = "200"
    )
}
