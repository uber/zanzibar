namespace java com.uber.zanzibar.clientless

include "../models/meta.thrift"

struct Request {
	1: optional string firstName
	2: optional string lastName
}

struct Response {
	1: optional string firstName
	2: optional string lastName1
}

service Clientless {
	Response beta(
		1: optional Request request
		2: optional string alpha
	) throws (
	) (
		zanzibar.http.method = "POST"
		zanzibar.http.reqHeaders = "x-uuid"
		zanzibar.http.path = "/clientless/post-request"
		zanzibar.http.status = "200"
		zanzibar.http.resHeaders = "x-uuid"
	)
	 // TODO: support headers annotation
     Response clientlessArgWithHeaders (
         1: required string name (
         zanzibar.http.ref = "headers.name"
            go.tag = "json:\"-\""
         )
         2: optional string userUUID (
         zanzibar.http.ref = "headers.x-uuid"
             go.tag = "json:\"-\""
         )
     ) (
         zanzibar.http.method = "POST"
         zanzibar.http.reqHeaders = "x-uuid"
         zanzibar.http.path = "/clientless/argWithHeaders"
         zanzibar.http.req.metadata = "meta.UUIDOnly"
         zanzibar.http.res.metadata = "meta.Grault"
         zanzibar.http.status = "200"
     )

     void emptyclientlessRequest(
        1: optional string testString
     ) (
         zanzibar.http.method = "GET"
         zanzibar.http.path = "/clientless/emptyclientlessRequest"
         zanzibar.http.status = "200"
     )
} (
)
