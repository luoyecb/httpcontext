package httpcontext

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type ContextFunc func(*Context)
type FilterFunc func(*Context)

// ========================================
// Context
// ========================================

type Context struct {
	*http.Request
	http.ResponseWriter

	Query       url.Values
	bodyBytes   []byte
	readBodyErr error

	values     map[string]interface{}
	pathParams map[string]string

	filters     []FilterFunc
	filterIndex int

	logger Logger
}

func NewContext(resp http.ResponseWriter, req *http.Request, filters ...FilterFunc) *Context {
	c := &Context{
		Request:        req,
		ResponseWriter: resp,
		Query:          req.URL.Query(),
		values:         map[string]interface{}{},
		logger:         &defaultLogger{},
	}
	if len(filters) > 0 {
		c.filters = append(c.filters, filters...)
	}

	if err := req.ParseForm(); err != nil {
		c.Write500(err)
		c.logger.LogError("ParseForm failed", err)
		return nil
	}

	return c
}

func (c *Context) Req() *http.Request {
	return c.Request
}

func (c *Context) RespWriter() http.ResponseWriter {
	return c.ResponseWriter
}

func (c *Context) SetLogger(l Logger) {
	c.logger = l
}

func (c *Context) Scheme() string {
	return c.Request.URL.Scheme // Cannot get scheme
}

func (c *Context) Host() string {
	return c.Request.Host
}

func (c *Context) RawQuery() string {
	return c.Request.URL.RawQuery
}

func (c *Context) Path() string {
	return c.Request.URL.Path
}

func (c *Context) RemoteAddr() string {
	return c.Request.RemoteAddr
}

func (c *Context) RequestURI() string {
	return c.Request.RequestURI
}

func (c *Context) Method() string {
	return c.Request.Method
}

func (c *Context) Header() http.Header {
	return c.Request.Header
}

// Return GET param
func (c *Context) Get(name string) string {
	return c.Query.Get(name)
}

func (c *Context) GetSlice(name string) []string {
	return c.Query[name]
}

// Return POST|PUT param
func (c *Context) Post(name string) string {
	return c.Request.PostFormValue(name)
}

func (c *Context) PostSlice(name string) []string {
	return c.Request.PostForm[name]
}

// Return POST|PUT|GET param
func (c *Context) Param(name string) string {
	return c.Request.FormValue(name)
}

func (c *Context) ParamSlice(name string) []string {
	return c.Request.Form[name]
}

func (c *Context) Params() url.Values {
	return c.Request.Form
}

func (c *Context) PathParam(name string) string {
	if c.pathParams != nil {
		return c.pathParams[name]
	}
	return ""
}

func (c *Context) PathParams() map[string]string {
	return c.pathParams
}

func (c *Context) readBody() {
	if c.readBodyErr != nil {
		return
	}
	if c.bodyBytes != nil {
		return
	}

	if body, err := ioutil.ReadAll(c.Request.Body); err != nil {
		c.readBodyErr = err
		c.logger.LogError("Read req body failed", err)
	} else {
		c.bodyBytes = body
	}
}

func (c *Context) BodyAsBytes() []byte {
	c.readBody()
	return c.bodyBytes
}

func (c *Context) BodyAsString() string {
	c.readBody()
	return string(c.bodyBytes)
}

func (c *Context) BodyAsJson(v interface{}) error {
	c.readBody()
	if err := json.Unmarshal(c.bodyBytes, v); err != nil {
		c.logger.LogError("Body Unmarshal json failed", err)
		return err
	}
	return nil
}

func (c *Context) WriteStatus(status int) {
	c.ResponseWriter.WriteHeader(status)
}

func (c *Context) Write404() {
	c.WriteStatus(http.StatusNotFound)
}

func (c *Context) Write500(err error) {
	c.WriteStatus(http.StatusInternalServerError)
	c.Write(err.Error())
}

func (c *Context) WriteHeader(key string, val string) {
	c.ResponseWriter.Header().Set(key, val)
}

func (c *Context) Write(format string, v ...interface{}) {
	fmt.Fprintf(c.ResponseWriter, format, v...)
}

func (c *Context) WriteString(data string) {
	c.WriteHeader("Content-Type", "text/plain")
	c.Write(data)
}

func (c *Context) WriteBytes(bytes []byte) {
	c.WriteString(string(bytes))
}

func (c *Context) WriteJson(v interface{}) {
	if jsBytes, err := json.Marshal(v); err != nil {
		c.logger.LogError("json.Marshal failed", err)
		c.Write500(err)
	} else {
		c.WriteHeader("Content-Type", "application/json")
		c.Write(string(jsBytes))
	}
}

func (c *Context) PutValue(key string, v interface{}) {
	c.values[key] = v
}

func (c *Context) GetValue(key string) interface{} {
	return c.values[key]
}

// ========================================
// Filter
// ========================================

func (c *Context) AppendFilters(filters ...FilterFunc) {
	if len(filters) > 0 {
		c.filters = append(c.filters, filters...)
	}
}

func (c *Context) processWithFilters(fn ContextFunc) {
	defer func() {
		if v := recover(); v != nil {
			if err, ok := v.(error); ok {
				c.logger.LogError("Recoverd", err)
				c.Write500(err)
			} else {
				c.logger.LogError("Recoverd", nil)
				c.Write500(nil)
			}
		}
	}()

	c.AppendFilters((FilterFunc)(fn))
	c.execFilters()
}

func (c *Context) execFilters() {
	for l := len(c.filters); c.filterIndex < l; c.filterIndex++ {
		fn := c.filters[c.filterIndex]
		fn(c)
	}
}

func (c *Context) Next() {
	c.filterIndex++
	c.execFilters()
}

func (c *Context) Abort() {
	c.filterIndex = len(c.filters) + 1
}
