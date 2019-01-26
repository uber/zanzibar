// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package zanzibar

import (
    "context"

    "go.uber.org/thriftrw/wire"
)

// AdapterTchannelStack is a stack of Adapter Handlers that can be invoked as an Handle.
// AdapterTchannelStack adapters are evaluated for requests in the order that they are added to the stack
// followed by the underlying HandlerFn. The adapter responses are then executed in reverse.
type AdapterTchannelStack struct {
    adapters        []AdapterTchannelHandle
    tchannelHandler TChannelHandler
}

// NewAdapterTchannelStack returns a new AdapterStack instance with no adapter preconfigured.
func NewAdapterTchannelStack(adapters []AdapterTchannelHandle,
    handler TChannelHandler) *AdapterTchannelStack {
    return &AdapterTchannelStack{
        tchannelHandler: handler,
        adapters:        adapters,
    }
}

// TchannelAdapters returns a list of all the handlers in the current AdapterStack.
func (m *AdapterTchannelStack) TchannelAdapters() []AdapterTchannelHandle {
    return m.adapters
}

// AdapterTchannelHandle used to define adapter
type AdapterTchannelHandle interface {
    // implement HandleRequest for your adapter. Return false
    // if the handler writes to the response body.
    HandleRequest(
        ctx context.Context,
        reqHeaders map[string]string,
        wireValue *wire.Value) (bool, error)

    // implement HandleResponse for your adapter. Return false
    // if the handler writes to the response body.
    HandleResponse(
        ctx context.Context,
        rwt RWTStruct) RWTStruct
}

// Handle executes the adapters in a stack and underlying handler.
func (m *AdapterTchannelStack) Handle(
    ctx context.Context,
    reqHeaders map[string]string,
    wireValue *wire.Value) (bool, RWTStruct, map[string]string, error) {
    var res RWTStruct
    var ok bool

    for i := 0; i < len(m.adapters); i++ {
        ok, err := m.adapters[i].HandleRequest(ctx, reqHeaders, wireValue)
        if ok == false {
            return ok, nil, map[string]string{}, err
        }
    }

    ok, res, resHeaders, err := m.tchannelHandler.Handle(ctx, reqHeaders, wireValue)
    for i := len(m.adapters) - 1; i >= 0; i-- {
        res = m.adapters[i].HandleResponse(ctx, res)
    }

    return ok, res, resHeaders, err
}
