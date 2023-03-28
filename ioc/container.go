package ioc

import (
	"reflect"
	"strings"
	"sync"
)

type implementationDetails struct {
	value        interface{}
	dependencies []reflect.Type
}

var implMap = make(map[string]implementationDetails)
var implMapMu = sync.RWMutex{}

func RegisterImpl[T any](creator interface{}) {
	var t *T = nil
	tt := reflect.TypeOf(t)
	tt = tt.Elem()

	// check that creator is a function
	reflectedCreator := reflect.ValueOf(creator)
	if reflectedCreator.Kind() != reflect.Func {
		panic("creator must be a function")
	}

	// check that creator returns a pointer to T
	if reflectedCreator.Type().NumOut() != 1 {
		panic("creator must return a single value")
	}

	// check that creator returns a pointer to T or T must be an interface
	if reflectedCreator.Type().Out(0).Kind() != reflect.Ptr && tt.Kind() != reflect.Interface {
		panic("creator must return a pointer to T or T must be an interface")
	}

	// check that T is not a pointer to an interface
	if tt.Kind() == reflect.Ptr && tt.Elem().Kind() == reflect.Interface {
		panic("T must not be a pointer to an interface")
	}

	// get list of parameters from creator
	creatorParams := make([]reflect.Type, reflectedCreator.Type().NumIn())
	for i := 0; i < reflectedCreator.Type().NumIn(); i++ {
		creatorParams[i] = reflectedCreator.Type().In(i)
	}

	arguments := make([]reflect.Value, len(creatorParams))
	for i, param := range creatorParams {
		// check that parameter is registered
		if impl, ok := implMap[param.String()]; ok {
			arguments[i] = reflect.ValueOf(impl.value)
		} else {
			panic("parameter " + param.String() + " is not registered")
		}
	}

	implMapMu.Lock()
	defer implMapMu.Unlock()

	// call creator with registered parameters
	impl := reflectedCreator.Call(arguments)[0].Interface()

	// if tt is an interface, check that value is a pointer to a struct
	if tt.Kind() == reflect.Interface && reflect.TypeOf(impl).Kind() != reflect.Ptr {
		panic("if T is an interface, creator must return a pointer to a struct")
	}

	implMap[tt.String()] = implementationDetails{
		value:        impl,
		dependencies: creatorParams,
	}

	// if tt is an interface and value, register value as a pointer to a struct
	if tt.Kind() == reflect.Interface {
		implMap[reflect.TypeOf(impl).String()] = implementationDetails{
			value:        impl,
			dependencies: creatorParams,
		}
	}
}

func GetImpl[T any]() T {
	var t *T = nil
	tt := reflect.TypeOf(t)
	tt = tt.Elem()

	implMapMu.RLock()
	defer implMapMu.RUnlock()

	impl, ok := implMap[tt.String()]
	if !ok {
		panic("implementation for " + tt.String() + " is not registered")
	}

	return impl.value.(T)
}

func GenerateDependencyGraph() string {
	implMapMu.RLock()
	defer implMapMu.RUnlock()

	graph := make([]string, 0)
	graph = append(graph, "digraph G {")

	seenTypes := make(map[string]bool, 0)

	for _, impl := range implMap {
		for _, dep := range impl.dependencies {
			if _, ok := seenTypes[reflect.TypeOf(impl.value).String()]; !ok {
				seenTypes[reflect.TypeOf(impl.value).String()] = true
			}

			if _, ok := seenTypes[dep.String()]; !ok {
				seenTypes[dep.String()] = true
			}

			graph = append(graph, "\t\""+reflect.TypeOf(impl.value).String()+"\" -> \""+dep.String()+"\";")
		}
	}

	// connect interfaces to their implementations with dashed lines
	for k, impl := range implMap {
		if reflect.TypeOf(impl.value).Kind() == reflect.Ptr && k != reflect.TypeOf(impl.value).String() {
			graph = append(graph, "\t\""+k+"\" -> \""+reflect.TypeOf(impl.value).String()+"\" [style=dashed];")
		}
	}

	// remove duplicate edges
	seen := make(map[string]bool)
	deduped := make([]string, 0)

	for _, edge := range graph {
		if _, ok := seen[edge]; !ok {
			seen[edge] = true
			deduped = append(deduped, edge)
		}
	}

	deduped = append(deduped, "}")
	return strings.Join(deduped, "\n") + "\n"
}
