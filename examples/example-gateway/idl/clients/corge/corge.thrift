namespace java com.uber.zanzibar.clients.corge

service Corge {
    string echoString(
        1: required string arg
    )
    // this method is intentionally not exposed in client-config.json
    bool echoBool(
        1: required bool arg
    )
}
