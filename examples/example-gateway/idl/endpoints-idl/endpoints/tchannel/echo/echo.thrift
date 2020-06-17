namespace java com.uber.zanzibar.endpoint.echo

service Echo {
  string echo(
    1: required string msg
  )
}
