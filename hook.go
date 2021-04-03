package webutils

import (
	"fmt"
	"reflect"

	"github.com/maxence-charriere/go-app/v8/pkg/app"
)

type Hook interface {
	SetResult(interface{}) Hook
	SetError(*error) Hook
	Use(composer app.Composer, ctx app.Context)
}

type Action func(app.Context) (interface{}, error)

type baseHook struct {
	result interface{}
	error  interface{}
	act    Action
}

func NewHook(act Action) Hook {
	return &baseHook{
		result: nil,
		error:  nil,
		act:    act,
	}
}

func (hook *baseHook) SetResult(v interface{}) Hook {
	hook.result = hook.getPointer(v)
	return hook
}

func (hook *baseHook) SetError(err *error) Hook {
	hook.error = hook.getPointer(err)
	return hook
}

func (hook *baseHook) Use(composer app.Composer, ctx app.Context) {
	ctx.Async(func() {
		result, err := hook.act(ctx)
		ctx.Dispatch(func() {
			hook.indirectSet(hook.result, result)
			hook.indirectSet(hook.error, err)
			if err != nil {
				fmt.Println(err)
			}
			composer.Update()
		})
	})
}

func (hook *baseHook) indirectSet(field interface{}, value interface{}) {
	if field == nil {
		return
	}

	target := reflect.Indirect(reflect.ValueOf(field))
	if value == nil {
		target.Set(reflect.Zero(target.Type()))
		return
	}

	v := reflect.ValueOf(value)
	if !v.Type().AssignableTo(target.Type()) {
		panic(fmt.Errorf("target type (%s) is not assignable to %s", target.Type(), v.Type()))
	}
	target.Set(v)
}

func (baseHook) getPointer(v interface{}) interface{} {
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Ptr {
		return v
	}
	return reflect.New(vv.Type()).Interface()
}
