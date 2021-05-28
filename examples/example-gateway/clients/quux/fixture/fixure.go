package fixture

import (
	mc "github.com/uber/zanzibar/examples/example-gateway/build/clients/quux/mock-client"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/foo/base/base"
)

var message = &base.Message{Body: "hola"}

// Fixture ...
var Fixture = &mc.ClientFixture{
	EchoString: &mc.EchoStringScenarios{
		Success: &mc.EchoStringFixture{
			Arg0: "echo",
			Ret0: "echo",
		},
	},
	EchoMessage: &mc.EchoMessageScenarios{
		Success: &mc.EchoMessageFixture{
			Arg0: message,
			Ret0: message,
		},
	},
}
