package ioc

import (
	"reflect"
	"strings"
	"sync"
)

type dependencyKind int

const (
	singleton dependencyKind = iota
	generator
)

type implementationDetails struct {
	value        interface{}
	dependencies []reflect.Type
	kind         dependencyKind
}

type Container struct {
	implMap   map[string]implementationDetails
	implMapMu sync.RWMutex
}

func NewContainer() *Container {
	return &Container{
		implMap:   make(map[string]implementationDetails),
		implMapMu: sync.RWMutex{},
	}
}

func checkCreatorFunc(tt reflect.Type, reflectedCreator reflect.Value) {
	// check that creator is a function
	if reflectedCreator.Kind() != reflect.Func {
		panic("creator must be a function")
	}

	// check that creator returns a single value
	if reflectedCreator.Type().NumOut() != 1 {
		panic("creator must return a single value")
	}

	//RegisterSingleton[domain.Application](func () domain.Application -> Valid
	//RegisterSingleton[*domain.ApplicationImpl](func () *domain.ApplicationImpl -> Valid
	//RegisterSingleton[domain.Application](func () *domain.ApplicationImpl -> Valid
	//RegisterSingleton[domain.Application](func () *domain.Application -> Invalid
	//RegisterSingleton[domain.Application](func () domain.ApplicationImpl -> Invalid
	//RegisterSingleton[domain.ApplicationImpl](func () domain.ApplicationImpl -> Invalid
	//RegisterSingleton[*domain.ApplicationImpl](func () domain.ApplicationImpl -> Invalid
	//RegisterSingleton[*domain.ApplicationImpl](func () *domain.Application -> Invalid
	//RegisterSingleton[*domain.ApplicationImpl](func () domain.Application -> Invalid

	// type paramater has to be an interface or a pointer to a struct
	if !(tt.Kind() == reflect.Interface) &&
		!(tt.Kind() == reflect.Ptr && tt.Elem().Kind() == reflect.Struct) {
		panic("T must be either an interface or a pointer to a struct")
	}

	// if creator returns a pointer, check that it is a pointer to a struct
	if reflectedCreator.Type().Out(0).Kind() == reflect.Ptr && reflectedCreator.Type().Out(0).Elem().Kind() != reflect.Struct {
		panic("if creator returns a pointer, it must be a pointer to a struct")
	}

	// if T is a pointer to a struct, check that the returned value is a pointer to a struct and that the returned value is equal to T
	if tt.Kind() == reflect.Ptr && tt.Elem().Kind() == reflect.Struct && (reflectedCreator.Type().Out(0).Kind() != reflect.Ptr || reflectedCreator.Type().Out(0).Elem().Kind() != reflect.Struct || tt != reflectedCreator.Type().Out(0)) {
		panic("T must be equal to the type of the returned value if T is a pointer to a struct")
	}

	// if T is an interface, check that the returned value implements it
	// or if creator returns a pointer to a struct, check that the underlying struct implements T
	// or T is equal to the returned value
	if !(tt.Kind() == reflect.Interface && reflectedCreator.Type().Out(0).Implements(tt)) &&
		!(tt.Kind() == reflect.Interface && reflectedCreator.Type().Out(0).Kind() == reflect.Ptr && reflectedCreator.Type().Out(0).Elem().Implements(tt)) &&
		tt != reflectedCreator.Type().Out(0) {
		panic("T must be equal to the type of the returned value or the returned value must implement T")
	}
}

func getCreatorParams(reflectedCreator reflect.Value) []reflect.Type {

	// get list of parameters from creator
	creatorParams := make([]reflect.Type, reflectedCreator.Type().NumIn())
	for i := 0; i < reflectedCreator.Type().NumIn(); i++ {
		creatorParams[i] = reflectedCreator.Type().In(i)
	}

	return creatorParams
}

func generateFromCreator(c *Container, tt reflect.Type, reflectedCreator reflect.Value) (interface{}, []reflect.Type) {
	// get list of parameters from creator
	creatorParams := getCreatorParams(reflectedCreator)

	arguments := make([]reflect.Value, len(creatorParams))
	for i, param := range creatorParams {
		// check that parameter is registered
		if impl, ok := c.implMap[param.String()]; ok {
			arguments[i] = reflect.ValueOf(impl.value)
		} else {
			panic("parameter " + param.String() + " is not registered")
		}
	}

	// call creator with registered parameters
	impl := reflectedCreator.Call(arguments)[0].Interface()

	// check that the returned value is not nil
	if reflect.ValueOf(impl).IsNil() {
		panic("creator must not return nil")
	}

	return impl, creatorParams
}

func RegisterSingleton[T any](c *Container, creator interface{}) {
	var t *T = nil
	tt := reflect.TypeOf(t)
	tt = tt.Elem()

	c.implMapMu.Lock()
	defer c.implMapMu.Unlock()

	reflectedCreator := reflect.ValueOf(creator)
	checkCreatorFunc(tt, reflectedCreator)
	impl, creatorParams := generateFromCreator(c, tt, reflectedCreator)

	c.implMap[tt.String()] = implementationDetails{
		value:        impl,
		dependencies: creatorParams,
		kind:         singleton,
	}

	// if tt is an interface and value, register value as a pointer to a struct
	if tt.Kind() == reflect.Interface {
		c.implMap[reflect.TypeOf(impl).String()] = implementationDetails{
			value:        impl,
			dependencies: creatorParams,
			kind:         singleton,
		}
	}
}

func RegisterGenerator[T any](c *Container, creator interface{}) {
	var t *T = nil
	tt := reflect.TypeOf(t)
	tt = tt.Elem()

	c.implMapMu.Lock()
	defer c.implMapMu.Unlock()

	reflectedCreator := reflect.ValueOf(creator)

	checkCreatorFunc(tt, reflectedCreator)
	creatorParams := getCreatorParams(reflectedCreator)

	c.implMap[tt.String()] = implementationDetails{
		value:        creator,
		dependencies: creatorParams,
		kind:         generator,
	}
}

func getFromType(c *Container, tt reflect.Type) interface{} {
	c.implMapMu.RLock()
	defer c.implMapMu.RUnlock()

	impl, ok := c.implMap[tt.String()]
	if !ok {
		panic("implementation for " + tt.String() + " is not registered")
	}

	if impl.kind == singleton {
		return impl.value
	}

	actual, _ := generateFromCreator(c, tt, reflect.ValueOf(impl.value))
	return actual
}

func Get[T any](c *Container) T {
	var t *T = nil
	tt := reflect.TypeOf(t)
	tt = tt.Elem()

	return getFromType(c, tt).(T)
}

func ForStruct[T any](c *Container) *T {
	var t *T = new(T)
	tt := reflect.TypeOf(t)
	tt = tt.Elem()

	c.implMapMu.RLock()
	defer c.implMapMu.RUnlock()

	// check that tt is a struct
	if tt.Kind() != reflect.Struct {
		panic("T must be a struct")
	}

	// get list of fields from struct
	reflectedStruct := reflect.ValueOf(t)
	reflectedStruct = reflectedStruct.Elem()
	structFields := make([]reflect.StructField, reflectedStruct.NumField())
	for i := 0; i < reflectedStruct.NumField(); i++ {
		structFields[i] = reflectedStruct.Type().Field(i)
	}

	// check that all fields without `inject:"ignore"` tag are registered
	for _, field := range structFields {
		if field.Tag.Get("inject") != "ignore" {
			if _, ok := c.implMap[field.Type.String()]; !ok {
				panic("field " + field.Name + " is not registered")
			}
		}
	}

	// inject fields with getFromType
	for i, field := range structFields {
		if field.Tag.Get("inject") != "ignore" {
			reflectedStruct.Field(i).Set(reflect.ValueOf(getFromType(c, field.Type)))
			continue
		}

		// if field is ignored, set it to zero value
		reflectedStruct.Field(i).Set(reflect.Zero(field.Type))
	}

	return t
}

func GenerateDependencyGraph(c *Container) string {
	c.implMapMu.RLock()
	defer c.implMapMu.RUnlock()

	graph := make([]string, 0)
	graph = append(graph, "digraph G {")

	seenTypes := make(map[string]bool, 0)

	for _, impl := range c.implMap {
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
	for k, impl := range c.implMap {
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
