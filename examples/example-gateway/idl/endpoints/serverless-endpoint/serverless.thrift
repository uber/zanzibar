namespace java com.uber.zanzibar.serverless

struct Request {
	1: optional string firstName
}

struct Response {
	1: optional string firstName
}

service Serverless {
	Response beta(
		1: required  Request request
	) throws (
	) (
		zanzibar.http.method = "POST"
		zanzibar.http.path = "/serverless/post-request/"
		zanzibar.http.status = "200"
	)
} (
)
