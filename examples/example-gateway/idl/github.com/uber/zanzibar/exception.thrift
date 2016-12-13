namespace java com.zanzibar.models.exception

enum BadRequestCode {
  zanzibar__bad_request
}

enum UnauthenticatedCode {
  zanzibar__unauthorized
}

enum UnauthorizedCode {
  zanzibar__forbidden
}

enum NotFoundCode {
  zanzibar__not_found
}

enum RateLimitedCode {
  zanzibar__too_many_requests
}

enum PayloadLimitedCode {
  zanzibar__payload_too_large
}

enum PermissionDeniedCode {
  zanzibar__permission_denied
}

enum InternalServerErrorCode {
  zanzibar__internal_server_error
}

enum NotAvailableCode {
  zanzibar__users__not_available
}

enum TemporaryRedirectCode {
  zanzibar__datacenter_redirect
}

exception TemporaryRedirect {
  1: required string location (zanzibar__http.ref = "headers.location")
  2: required TemporaryRedirectCode code
  3: optional string messageType
  4: optional string uri
}

exception BadRequest {
  1: required BadRequestCode code
  2: required string message
}

exception Unauthenticated {
  1: required UnauthenticatedCode code
  2: required string message
  3: optional string errorCode
}

exception Unauthorized {
  1: required UnauthorizedCode code
  2: required string message
}

exception NotFound {
  1: required NotFoundCode code
  2: required string message
}

exception RateLimited {
  1: required RateLimitedCode code
  2: required string message
}

exception PayloadLimited {
  1: required PayloadLimitedCode code
  2: required string message
}

exception PermissionDenied {
  1: required PermissionDeniedCode code
  2: required string message
}

exception ServerError {
  1: required InternalServerErrorCode code
  2: required string message
}

exception NotAvailable {
  1: required string message
  2: required NotAvailableCode code
}

enum UserBannedCode {
  zanzibar__users__account_banned
}

exception UserBanned {
  1: required UserBannedCode code
  2: required string message
}
