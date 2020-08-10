// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package twirp

import (
	"context"
	"encoding/json"
	"net/http"
)

// ServerHooks is a container for callbacks that can instrument a
// Twirp-generated server. These callbacks all accept a context and return a
// context. They can use this to add to the request context as it threads
// through the system, appending values or deadlines to it.
//
// The RequestReceived and RequestRouted hooks are special: they can return
// errors. If they return a non-nil error, handling for that request will be
// stopped at that point. The Error hook will be triggered, and the error will
// be sent to the client. This can be used for stuff like auth checks before
// deserializing a request.
//
// The RequestReceived hook is always called first, and it is called for every
// request that the Twirp server handles. The last hook to be called in a
// request's lifecycle is always ResponseSent, even in the case of an error.
//
// Details on the timing of each hook are documented as comments on the fields
// of the ServerHooks type.
type ServerHooks struct {
	// RequestReceived is called as soon as a request enters the Twirp
	// server at the earliest available moment.
	RequestReceived func(context.Context) (context.Context, error)

	// RequestRouted is called when a request has been routed to a
	// particular method of the Twirp server.
	RequestRouted func(context.Context) (context.Context, error)

	// ResponsePrepared is called when a request has been handled and a
	// response is ready to be sent to the client.
	ResponsePrepared func(context.Context) context.Context

	// ResponseSent is called when all bytes of a response (including an error
	// response) have been written. Because the ResponseSent hook is terminal, it
	// does not return a context.
	ResponseSent func(context.Context)

	// Error hook is called when an error occurs while handling a request. The
	// Error is passed as argument to the hook.
	Error func(context.Context, Error) context.Context
}

// CallRequestReceived call twirp.ServerHooks.RequestReceived if the hook is available
func (h *ServerHooks) CallRequestReceived(ctx context.Context) (context.Context, error) {
	if h == nil || h.RequestReceived == nil {
		return ctx, nil
	}
	return h.RequestReceived(ctx)
}

// CallRequestRouted call twirp.ServerHooks.RequestRouted if the hook is available
func (h *ServerHooks) CallRequestRouted(ctx context.Context) (context.Context, error) {
	if h == nil || h.RequestRouted == nil {
		return ctx, nil
	}
	return h.RequestRouted(ctx)
}

// CallResponsePrepared call twirp.ServerHooks.ResponsePrepared if the hook is available
func (h *ServerHooks) CallResponsePrepared(ctx context.Context) context.Context {
	if h == nil || h.ResponsePrepared == nil {
		return ctx
	}
	return h.ResponsePrepared(ctx)
}

// CallResponseSent call twirp.ServerHooks.ResponseSent if the hook is available
func (h *ServerHooks) CallResponseSent(ctx context.Context) {
	if h == nil || h.ResponseSent == nil {
		return
	}
	h.ResponseSent(ctx)
}

// CallError call twirp.ServerHooks.Error if the hook is available
func (h *ServerHooks) CallError(ctx context.Context, err Error) context.Context {
	if h == nil || h.Error == nil {
		return ctx
	}
	return h.Error(ctx, err)
}

// WriteError writes Twirp errors in the response and triggers hooks.
func (h *ServerHooks) WriteError(ctx context.Context, resp http.ResponseWriter, err error) {
	// Non-twirp errors are wrapped as Internal (default)
	twerr, ok := err.(Error)
	if !ok {
		twerr = InternalErrorWith(err)
	}

	statusCode := ServerHTTPStatusFromErrorCode(twerr.Code())
	ctx = WithStatusCode(ctx, statusCode)
	ctx = h.CallError(ctx, twerr)

	resp.Header().Set("Content-Type", "application/json") // Error responses are always JSON (instead of protobuf)
	resp.WriteHeader(statusCode)                          // HTTP response status code

	respBody := marshalErrorToJSON(twerr)
	_, writeErr := resp.Write(respBody)
	if writeErr != nil {
		// We have three options here. We could log the error, call the Error
		// hook, or just silently ignore the error.
		//
		// Logging is unacceptable because we don't have a user-controlled
		// logger; writing out to stderr without permission is too rude.
		//
		// Calling the Error hook would confuse users: it would mean the Error
		// hook got called twice for one request, which is likely to lead to
		// duplicated log messages and metrics, no matter how well we document
		// the behavior.
		//
		// Silently ignoring the error is our least-bad option. It's highly
		// likely that the connection is broken and the original 'err' says
		// so anyway.
		_ = writeErr
	}

	h.CallResponseSent(ctx)
}

// marshalErrorToJSON returns JSON from a twirp.Error, that can be used as HTTP error response body.
// If serialization fails, it will use a descriptive Internal error instead.
func marshalErrorToJSON(twerr Error) []byte {
	// make sure that msg is not too large
	msg := twerr.Msg()
	if len(msg) > 1e6 {
		msg = msg[:1e6]
	}

	type twerrJSON struct {
		Code string            `json:"code"`
		Msg  string            `json:"msg"`
		Meta map[string]string `json:"meta,omitempty"`
	}

	tj := twerrJSON{
		Code: string(twerr.Code()),
		Msg:  msg,
		Meta: twerr.MetaMap(),
	}

	buf, err := json.Marshal(&tj)
	if err != nil {
		buf = []byte("{\"type\": \"" + Internal + "\", \"msg\": \"There was an error but it could not be serialized into JSON\"}") // fallback
	}

	return buf
}

// ChainHooks creates a new *ServerHooks which chains the callbacks in
// each of the constituent hooks passed in. Each hook function will be
// called in the order of the ServerHooks values passed in.
//
// For the erroring hooks, RequestReceived and RequestRouted, any returned
// errors prevent processing by later hooks.
func ChainHooks(hooks ...*ServerHooks) *ServerHooks {
	if len(hooks) == 0 {
		return nil
	}
	if len(hooks) == 1 {
		return hooks[0]
	}
	return &ServerHooks{
		RequestReceived: func(ctx context.Context) (context.Context, error) {
			var err error
			for _, h := range hooks {
				if h != nil && h.RequestReceived != nil {
					ctx, err = h.RequestReceived(ctx)
					if err != nil {
						return ctx, err
					}
				}
			}
			return ctx, nil
		},
		RequestRouted: func(ctx context.Context) (context.Context, error) {
			var err error
			for _, h := range hooks {
				if h != nil && h.RequestRouted != nil {
					ctx, err = h.RequestRouted(ctx)
					if err != nil {
						return ctx, err
					}
				}
			}
			return ctx, nil
		},
		ResponsePrepared: func(ctx context.Context) context.Context {
			for _, h := range hooks {
				if h != nil && h.ResponsePrepared != nil {
					ctx = h.ResponsePrepared(ctx)
				}
			}
			return ctx
		},
		ResponseSent: func(ctx context.Context) {
			for _, h := range hooks {
				if h != nil && h.ResponseSent != nil {
					h.ResponseSent(ctx)
				}
			}
		},
		Error: func(ctx context.Context, twerr Error) context.Context {
			for _, h := range hooks {
				if h != nil && h.Error != nil {
					ctx = h.Error(ctx, twerr)
				}
			}
			return ctx
		},
	}
}
