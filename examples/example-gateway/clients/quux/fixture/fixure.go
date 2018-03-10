package fixture

import (
	mc "github.com/uber/zanzibar/examples/example-gateway/build/clients/quux/mock-client"
)

// Fixture ...
var Fixture = &mc.ClientFixture{
	EchoString: &mc.EchoStringScenarios{
		Success: &mc.EchoStringFixture{
			Arg0: "echo",
			Ret0: "echo",
		},
	},
}
