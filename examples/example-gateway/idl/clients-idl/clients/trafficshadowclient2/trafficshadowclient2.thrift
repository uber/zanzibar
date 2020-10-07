namespace java com.uber.zanzibar.clients.trafficshadowclient2

typedef string UUID
typedef i64 (json.type = 'Date') Timestamp
typedef i64 (json.type = "Long") Long
typedef list<string> StringList
typedef list<UUID> UUIDList

struct TrafficShadowClient2Response {
    1: required string resField2
}

exception TrafficShadowException {
    1: required string stringField (zanzibar.http.ref = "headers.another-header-field")
}

exception SeeOthersRedirection {
}

service Trafficshadowclient2 {
    TrafficShadowClient2Response helloWorld(
    ) throws (
        1: TrafficShadowException trafficShadowException (zanzibar.http.status = "403")
        2: SeeOthersRedirection seeOthersRedirection (zanzibar.http.status = "303", zanzibar.http.res.body.disallow = "true")
    ) (
       zanzibar.http.method = "GET"
       zanzibar.http.path = "/trafficshadow/hello"
       zanzibar.http.status = "200"
    )
}