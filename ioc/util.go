package ioc

import "reflect"

func checkCreatorFunc(tt reflect.Type, reflectedCreator reflect.Value) {
	// check that creator is a function
	if reflectedCreator.Kind() != reflect.Func {
		panic("creator must be a function")
	}

	// check that creator returns a single value
	if reflectedCreator.Type().NumOut() != 1 {
		panic("creator must return a single value")
	}

	//RegisterSingleton[domain.Exhibit](func () domain.Exhibit -> Valid
	//RegisterSingleton[*domain.ExhibitImpl](func () *domain.ExhibitImpl -> Valid
	//RegisterSingleton[domain.Exhibit](func () *domain.ExhibitImpl -> Valid
	//RegisterSingleton[domain.Exhibit](func () *domain.Exhibit -> Invalid
	//RegisterSingleton[domain.Exhibit](func () domain.ExhibitImpl -> Invalid
	//RegisterSingleton[domain.ExhibitImpl](func () domain.ExhibitImpl -> Invalid
	//RegisterSingleton[*domain.ExhibitImpl](func () domain.ExhibitImpl -> Invalid
	//RegisterSingleton[*domain.ExhibitImpl](func () *domain.Exhibit -> Invalid
	//RegisterSingleton[*domain.ExhibitImpl](func () domain.Exhibit -> Invalid

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

func generateFromCreator(c *Container, reflectedCreator reflect.Value) (any, []reflect.Type) {
	// get list of parameters from creator
	creatorParams := getCreatorParams(reflectedCreator)
	arguments := getDependencies(c, creatorParams)
	// call creator with registered parameters
	impl := reflectedCreator.Call(arguments)[0].Interface()

	// check that the returned value is not nil
	if reflect.ValueOf(impl).IsNil() {
		panic("creator must not return nil")
	}

	return impl, creatorParams
}

func getFromType(c *Container, tt reflect.Type) any {
	c.implMapMu.RLock()
	defer c.implMapMu.RUnlock()

	impl, ok := c.implMap[tt.String()]
	if !ok {
		panic("implementation for " + tt.String() + " is not registered")
	}

	if impl.kind == singleton {
		return impl.value
	}

	actual, _ := generateFromCreator(c, reflect.ValueOf(impl.value))
	return actual
}

func getDependencies(c *Container, params []reflect.Type) []reflect.Value {
	arguments := make([]reflect.Value, len(params))
	for i, param := range params {
		// check that parameter is registered
		if impl, ok := c.implMap[param.String()]; ok {
			arguments[i] = reflect.ValueOf(impl.value)
		} else {
			panic("parameter " + param.String() + " is not registered")
		}
	}

	return arguments
}

func checkForFunc(fn any) {
	// check that fn is a function
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		panic("fn must be a function")
	}

	// check that the function does not return anything
	if reflect.TypeOf(fn).NumOut() != 0 {
		panic("fn must not return anything")
	}
}

func forFuncGetArgumentsAndReflectedFn(c *Container, fn any) ([]reflect.Value, reflect.Value) {
	checkForFunc(fn)

	c.implMapMu.RLock()
	defer c.implMapMu.RUnlock()

	reflectedFn := reflect.ValueOf(fn)
	params := getCreatorParams(reflectedFn)
	arguments := getDependencies(c, params)

	return arguments, reflectedFn
}
