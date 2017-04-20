namespace java com.uber.zanzibar.clients.bar

include "../foo/foo.thrift"

struct BarRequest {
  1: required string stringField (zanzibar.http.ref = "params.someParamsField")
  2: required bool boolField (zanzibar.http.ref = "query.some-query-field")
}
struct BarResponse {
  1: required string stringField (
    zanzibar.http.ref = "headers.some-header-field"
    zanzibar.validation.type = "object,number"
  )
  2: required i32 intWithRange
  3: required i32 intWithoutRange (zanzibar.ignore.integer.range = "true")
  4: required map<string, i32> mapIntWithRange
  5: required map<string, i32> mapIntWithoutRange (zanzibar.ignore.integer.range = "true")
}

exception BarException {
  1: required string stringField (zanzibar.http.ref = "headers.another-header-field")
}

service Bar {
  BarResponse normal (
    1: required BarRequest request
  ) throws (
    1: BarException barException (zanzibar.http.status = "403")
  ) (
    zanzibar.http.method = "POST"
    zanzibar.http.path = "/bar/bar-path"
    zanzibar.http.status = "200"
    zanzibar.http.req.def.boxed = "true"
    zanzibar.meta = "SomeMeta"
    zanzibar.handler = "bar.baz"
  )
  BarResponse noRequest (
  ) throws (
    1: BarException barException (zanzibar.http.status = "403")
  ) (
    zanzibar.http.method = "GET"
    zanzibar.http.path = "/bar/no-request-path"
    zanzibar.http.status = "200"
    zanzibar.http.req.def.boxed = "false"
    zanzibar.meta = "SomeMeta"
    zanzibar.handler = "bar.baz"
  )
  BarResponse missingArg (
  ) throws (
    1: BarException barException (zanzibar.http.status = "403")
  ) (
    zanzibar.http.method = "GET"
    zanzibar.http.path = "/bar/missing-arg-path"
    zanzibar.http.status = "200"
    zanzibar.http.req.def.boxed = "true"
    zanzibar.meta = "SomeMeta"
    zanzibar.handler = "bar.baz"
  )
  BarResponse tooManyArgs (
    1: required BarRequest request
    2: optional foo.FooStruct foo
  ) throws (
    1: BarException barException (zanzibar.http.status = "403")
  ) (
    zanzibar.http.headers = "x-uuid,x-token"
    zanzibar.http.method = "POST"
    zanzibar.http.path = "/bar/too-many-args-path"
    zanzibar.http.status = "200"
    zanzibar.http.req.def.boxed = "true"
    zanzibar.meta = "SomeMeta"
    zanzibar.handler = "bar.baz"
    zanzibar.http.reqHeaders = "x-uuid,x-token"
    zanzibar.http.resHeaders = "x-uuid,x-token"
  )
  void argNotStruct (
    1: required string request
  ) throws (
    1: BarException barException (zanzibar.http.status = "403")
  ) (
    zanzibar.http.method = "POST"
    zanzibar.http.path = "/bar/arg-not-struct-path"
    zanzibar.http.status = "200"
    zanzibar.http.req.def.boxed = "false"
    zanzibar.meta = "SomeMeta"
    zanzibar.handler = "bar.baz"
  )
}
