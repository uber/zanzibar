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
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/uber-go/zap"
	"github.com/uber/tchannel-go"
)

// TChannelServer wraps a standard tchannel.Channel, but improves
// interoperability with the tchannel client and other dependencies related to
// zanzibar
type TChannelServer struct {
	*tchannel.Channel
	Logger zap.Logger
}

// TChannelServerOptions are used to initialize the TChannel wrapper struct
type TChannelServerOptions struct {
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

// NewTChannelServer allocates a new TChannel wrapper struct
func NewTChannelServer(opts *TChannelServerOptions) (*TChannelServer, error) {
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
		return nil, errors.Errorf(
			"Error creating TChannel Server:\n    %s",
			err)
	}

	return &TChannelServer{
		Channel: channel,
		Logger:  opts.Logger,
	}, nil
}
