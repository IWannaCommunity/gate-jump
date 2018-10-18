package res

import (
	"encoding/json"
	"errors"
	"gate-jump/util"
	"log"
	"net/http"
)

type Response struct {
	Function string
	Code     int
	Payload  struct {
		Success bool
		Error   *ResponseError
		Data    interface{}
	}
	InternalError *ServerError
}

type ResponseError struct {
	Code    int
	Message string
}

type ServerError struct {
	Err   error
	Args  []interface{}
	Query string
}

func New(code int) *Response {
	r := Response{}
	r.Function = util.MyCaller()
	r.Code = code
	r.Payload.Success = true
	r.InternalError = nil
	r.Payload.Error = nil
	return &r
}

func CreateInternalError(query string, args []interface{}, err error) *ServerError {
	return &ServerError{Query: query, Args: args, Err: err}
}

func (r *Response) SetData(data interface{}) *Response {
	r.Payload.Data = data
	return r
}
func (r *Response) SetToken(token string) *Response {
	r.Payload.Data = map[string]string{"token": token}
	return r
}
func (r *Response) SetErrorMessage(message string) *Response {
	e := ResponseError{}
	e.Code = r.Code
	e.Message = message
	r.Payload.Error = &e
	r.InternalError = &ServerError{Query: "", Args: nil, Err: errors.New(message)} // oddball case where code fails that isn't from sql query
	return r
}
func (r *Response) SetInternalError(serverError *ServerError) *Response {
	r.InternalError = serverError
	return r
}

func (r *Response) Error(w http.ResponseWriter) {
	if r.Code >= 500 {
		/*
			Example:
			2018/10/17 21:10:47
				Internal Server Error (500) in Function: getUser()
				"SELECT * FROM users WHERE id=?" << ({<nil>})
				Message: Invalid sql syntax check something something
		*/
		log.Printf("\n\tInternal Server Error (%d) in Function: %s()\n\t\"%s\" << %v\n\tMessage: %s",
			r.Code, r.Function, r.InternalError.Query, r.InternalError.Args, r.InternalError.Err.Error())
	}
	r.Payload.Success = false
	if r.Payload.Error == nil {
		e := ResponseError{}
		e.Code = r.Code
		e.Message = r.InternalError.Err.Error()
		r.Payload.Error = &e
	}
	r.JSON(w)
}

func (r *Response) JSON(w http.ResponseWriter) {
	p, _ := json.Marshal(r.Payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Code)
	w.Write(p)
}
