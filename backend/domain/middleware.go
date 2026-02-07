package domain

type Middleware interface {
	Execute(req *Request, next Handler) (*Response, error)
}

type MiddlewareFunc func(req *Request, next Handler) (*Response, error)

func (f MiddlewareFunc) Execute(req *Request, next Handler) (*Response, error) {
	return f(req, next)
}
