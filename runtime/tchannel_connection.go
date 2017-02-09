// Copyright (c) 2017 Uber Technologies, Inc.
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
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber-go/zap"
	"github.com/uber/tchannel-go"
)

// TChannelConnection wraps a stanadrd tchannel.Chanel, but improves
// interoperability with the tchannel client and other dependencies related to
// zanzibar
type TChannelConnection struct {
	*tchannel.Channel
	Logger zap.Logger
}

// TChannelConnectionOptions are used to initialize the TChannel wrapper struct
type TChannelConnectionOptions struct {
	DefaultConnectionOptions tchannel.ConnectionOptions
	Logger                   zap.Logger
	OnPeerStatusChanged      func(*tchannel.Peer)
	ProcessName              string
	RelayHost                tchannel.RelayHost
	RelayLocalHandlers       []string
	RelayMaxTimeout          time.Duration
	ServiceName              string
	StatsReporter            tchannel.StatsReporter
	Tracer                   opentracing.Tracer
}

// NewTChannel allocates a new TChannel wrapper struct
func NewTChannel(opts *TChannelConnectionOptions) (*TChannelConnection, error) {
	channel, err := tchannel.NewChannel(
		opts.ServiceName,
		&tchannel.ChannelOptions{
			DefaultConnectionOptions: opts.DefaultConnectionOptions,
			// TODO: Logger: opts.Logger,
			OnPeerStatusChanged: opts.OnPeerStatusChanged,
			ProcessName:         opts.ProcessName,
			RelayHost:           opts.RelayHost,
			RelayLocalHandlers:  opts.RelayLocalHandlers,
			RelayMaxTimeout:     opts.RelayMaxTimeout,
			StatsReporter:       opts.StatsReporter,
		})

	if err != nil {
		return nil, fmt.Errorf(
			"Error creating TChannel Connection:\n    %s",
			err.Error())
	}

	return &TChannelConnection{
		Channel: channel,
		Logger:  opts.Logger,
	}, nil
}
