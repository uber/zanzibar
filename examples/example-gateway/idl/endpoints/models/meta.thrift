namespace java com.uber.zanzibar.models.meta

/*
* Define header schema here
*
* different header models can be composed in endpoint annotation:
*  `zanzibar.http.req.metadata = "meta.Grault,meta.Fred"`
* */

struct Grault {
  1: optional string UUID (zanzibar.http.ref = "headers.X-Uuid")
  2: optional string token (zanzibar.http.ref = "headers.X-Token")
}

struct Garply {
  1: optional string UUID (zanzibar.http.ref = "headers.uuid")
  2: optional string token (zanzibar.http.ref = "headers.token")
}

struct Fred {
  1: required string contentType (zanzibar.http.ref = "headers.content-type")
  2: required string auth (zanzibar.http.ref = "headers.auth")
}

struct Thud {
  1: optional string someResHeader (zanzibar.http.ref = "headers.Some-Res-Header")
}

struct TokenOnly {
    1: required string Token (zanzibar.http.ref = "headers.X-Token")
}

struct UUIDOnly {
    1: required string UUID (zanzibar.http.ref = "headers.X-Uuid")
}

struct Dgx {
    1: required string s1 (zanzibar.http.ref = "headers.s1")
    2: required i32 i2 (zanzibar.http.ref = "headers.i2")
    3: optional bool b3 (zanzibar.http.ref = "headers.b3")
}
