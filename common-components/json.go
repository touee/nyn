package components

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/touee/nyn"
)

// HTTPResponseJSONDecorator 将得到的 HTTP 响应中的 body 解析
type HTTPResponseJSONDecorator struct{}

// DecoratePayload 转换 payload
func (HTTPResponseJSONDecorator) DecoratePayload(c *nyn.Crawler, task nyn.Task, payload interface{}) (decoratedPayload interface{}, err error) {
	var resp = payload.(*http.Response)
	defer resp.Body.Close()

	var dec = json.NewDecoder(resp.Body)
	var objPointer interface{}
	if x, ok := task.(interface{ GetPayloadType() reflect.Type }); ok {
		var t = x.GetPayloadType()
		if t.Kind() == reflect.Slice {

			var slice = reflect.MakeSlice(t, 0, 0)
			var x = reflect.New(slice.Type())
			x.Elem().Set(slice)
			objPointer = x.Interface()
		} else {
			objPointer = reflect.New(x.GetPayloadType()).Interface()
		}
	} else {
		objPointer = new(interface{})
	}
	if err = dec.Decode(objPointer); err != nil {
		return nil, err
	}
	return reflect.Indirect(reflect.ValueOf(objPointer)).Interface(), nil
}
