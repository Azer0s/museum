package ioc

import (
	"github.com/google/uuid"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type dependencyKind int

const (
	singleton dependencyKind = iota
	generator
)

type implementationDetails struct {
	value        any
	dependencies []reflect.Type
	kind         dependencyKind
}

type Container struct {
	implMap            map[string]implementationDetails
	implMapMu          sync.RWMutex
	funcDependencies   [][]reflect.Type
	funcDependenciesMu sync.RWMutex
	Debug              bool
}

func NewContainer() *Container {
	c := &Container{
		implMap:            make(map[string]implementationDetails),
		implMapMu:          sync.RWMutex{},
		funcDependencies:   make([][]reflect.Type, 0),
		funcDependenciesMu: sync.RWMutex{},
		Debug:              true,
	}
	RegisterSingleton[*Container](c, func() *Container {
		return c
	})
	return c
}

func RegisterSingleton[T any](c *Container, creator any) {
	var t *T = nil
	tt := reflect.TypeOf(t)
	tt = tt.Elem()

	c.implMapMu.Lock()
	defer c.implMapMu.Unlock()

	reflectedCreator := reflect.ValueOf(creator)
	checkCreatorFunc(tt, reflectedCreator)
	impl, creatorParams := generateFromCreator(c, reflectedCreator)

	c.implMap[tt.String()] = addDebugDependencies(c, implementationDetails{
		value: impl,
		kind:  singleton,
	}, creatorParams)

	// if tt is an interface and value, register value as a pointer to a struct
	if tt.Kind() == reflect.Interface {
		c.implMap[reflect.TypeOf(impl).String()] = addDebugDependencies(c, implementationDetails{
			value: impl,
			kind:  singleton,
		}, creatorParams)
	}
}

func RegisterGenerator[T any](c *Container, creator any) {
	var t *T = nil
	tt := reflect.TypeOf(t)
	tt = tt.Elem()

	c.implMapMu.Lock()
	defer c.implMapMu.Unlock()

	reflectedCreator := reflect.ValueOf(creator)

	checkCreatorFunc(tt, reflectedCreator)
	creatorParams := getCreatorParams(reflectedCreator)

	c.implMap[tt.String()] = addDebugDependencies(c, implementationDetails{
		value: creator,
		kind:  generator,
	}, creatorParams)
}

func Get[T any](c *Container) T {
	var t *T = nil
	tt := reflect.TypeOf(t)
	tt = tt.Elem()

	return getFromType(c, tt).(T)
}

func ForStruct[T any](c *Container) *T {
	var t = new(T)
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

func ForFunc(c *Container, fn any) {
	arguments, reflectedFn := forFuncGetArgumentsAndReflectedFn(c, fn)

	// call fn with registered parameters
	reflectedFn.Call(arguments)
}

func ForFuncAsync(c *Container, fn any) {
	arguments, reflectedFn := forFuncGetArgumentsAndReflectedFn(c, fn)

	// call fn with registered parameters in a goroutine
	go reflectedFn.Call(arguments)
}

func GenerateDependencyGraph(c *Container) string {
	if !c.Debug {
		panic("cannot generate dependency graph when debug is disabled")
	}

	c.implMapMu.RLock()
	defer c.implMapMu.RUnlock()

	c.funcDependenciesMu.RLock()
	defer c.funcDependenciesMu.RUnlock()

	graph := make([]string, 0)
	graph = append(graph, "digraph G {")

	seenTypes := make(map[string]bool)

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

	// connect anonymous functions to their dependencies
	for i, deps := range c.funcDependencies {
		funcId := uuid.New().String()
		graph = append(graph, "\t\""+funcId+"\" [label=\"func()/"+strconv.Itoa(i)+"\"];")
		for _, dep := range deps {
			if _, ok := seenTypes[dep.String()]; !ok {
				seenTypes[dep.String()] = true
			}

			graph = append(graph, "\t\""+funcId+"\" -> \""+dep.String()+"\";")
		}
	}

	// add struct pointer nodes that aren't used
	for _, impl := range c.implMap {
		if _, ok := seenTypes[reflect.TypeOf(impl.value).String()]; !ok {
			graph = append(graph, "\t\""+reflect.TypeOf(impl.value).String()+"\";")
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
