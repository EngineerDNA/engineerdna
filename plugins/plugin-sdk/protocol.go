package sdk

// This file contains protocol-level helpers for plugin communication

// NewRequest creates a new JSON-RPC request
func NewRequest(method string, params interface{}, id string) Request {
	return Request{
		Method: method,
		Params: params,
		ID:     id,
	}
}

// NewResponse creates a successful JSON-RPC response
func NewResponse(result interface{}, id string) Response {
	return Response{
		Result: result,
		ID:     id,
	}
}

// NewErrorResponse creates an error JSON-RPC response
func NewErrorResponse(code int, message string, id string) Response {
	return Response{
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
		ID: id,
	}
}
