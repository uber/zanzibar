package fixture

import (
	mc "github.com/uber/zanzibar/examples/selective-gateway/build/clients/echo/mock-client"
	gen "github.com/uber/zanzibar/examples/selective-gateway/build/proto-gen/clients/echo"
)

var echoEchoFixtures = &mc.EchoEchoScenarios{
	Success: &mc.EchoEchoFixture{
		Arg0Any: true,
		Arg1Any: true,
		Arg2Any: 1,

		Ret1: &gen.Response{Message: "hello"},
	},
}

// Fixture ...
var Fixture = &mc.ClientFixture{
	EchoEcho: echoEchoFixtures,
}
