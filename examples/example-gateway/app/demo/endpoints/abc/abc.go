package abc

import (
	"context"

	zanzibar "github.com/uber/zanzibar/runtime"

	module "github.com/uber/zanzibar/examples/example-gateway/build/app/demo/endpoints/abc/module"
	workflow "github.com/uber/zanzibar/examples/example-gateway/build/app/demo/endpoints/abc/workflow"
)

// NewAppDemoServiceCallWorkflow creates the demo app service callback workflow.
func NewAppDemoServiceCallWorkflow(deps *module.Dependencies) workflow.AppDemoServiceCallWorkflow {
	return &demo{}
}

type demo struct{}

func (h *demo) Handle(ctx context.Context, reqHeaders zanzibar.Header) (context.Context, int32, zanzibar.Header, error) {
	return ctx, 0, nil, nil
}
