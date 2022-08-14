package httpcontext

import (
	"net/http"
	"strings"
)

const (
	METHOD_GET    = "GET"
	METHOD_POST   = "POST"
	METHOD_PUT    = "PUT"
	METHOD_DELETE = "DELETE"
)

// ========================================
// ServeMux
// ========================================

type ServeMux struct {
	// mux     *http.ServeMux
	mux     *PathTree
	filters []FilterFunc // default filters
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		// mux: http.NewServeMux(),
		mux: NewPathTree(),
	}
}

func (m *ServeMux) AddFilters(filters ...FilterFunc) {
	m.filters = append(m.filters, filters...)
}

func (m *ServeMux) Handle(pattern string, fn ContextFunc, filters ...FilterFunc) {
	// m.mux.HandleFunc(pattern, HttpHandlerFunc(fn, copyFilters...))
	m.mux.Put(pattern, fn, filters...)
}

func (m *ServeMux) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// m.mux.ServeHTTP(resp, req)
	fn, filters, params := m.mux.FindHandler(req.URL.Path)
	if fn != nil {
		copyFilters := make([]FilterFunc, 0, len(m.filters)+len(filters))
		copyFilters = append(copyFilters, m.filters...)
		copyFilters = append(copyFilters, filters...)

		if ctx := NewContext(resp, req, copyFilters...); ctx != nil {
			ctx.pathParams = params
			ctx.processWithFilters(fn)
		}
	}
}

// ========================================
// Request handler
// ========================================

func HttpHandlerFunc(fn ContextFunc, filters ...FilterFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		if ctx := NewContext(resp, req, filters...); ctx != nil {
			ctx.processWithFilters(fn)
		}
	}
}

func Method(method string, fn ContextFunc, filters ...FilterFunc) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		if ctx := NewContext(resp, req, filters...); ctx != nil {
			if strings.ToUpper(method) == ctx.Method() {
				ctx.processWithFilters(fn)
			} else {
				ctx.Write404()
			}
		}
	}
}

func Get(fn ContextFunc, filters ...FilterFunc) http.HandlerFunc {
	return Method(METHOD_GET, fn, filters...)
}

func Post(fn ContextFunc, filters ...FilterFunc) http.HandlerFunc {
	return Method(METHOD_POST, fn, filters...)
}

func Put(fn ContextFunc, filters ...FilterFunc) http.HandlerFunc {
	return Method(METHOD_PUT, fn, filters...)
}

func Delete(fn ContextFunc, filters ...FilterFunc) http.HandlerFunc {
	return Method(METHOD_DELETE, fn, filters...)
}
