namespace java com.uber.zanzibar.endpoint.bounce

service Bounce {
  string bounce(
    1: required string msg
  )
}
