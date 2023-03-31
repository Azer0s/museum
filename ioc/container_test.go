package ioc

import (
	"fmt"
	"testing"
)

type Foo struct {
}

func NewFoo() *Foo {
	return &Foo{}
}

type Bar struct {
}

func NewBar() *Bar {
	return &Bar{}
}

type Bazzable interface {
	DoBazz()
}

type Bazz struct {
	Foo *Foo
	Bar *Bar
}

func (b Bazz) DoBazz() {
	fmt.Println("bazz")
}

func NewBazz(foo *Foo, bar *Bar) Bazzable {
	return &Bazz{
		Foo: foo,
		Bar: bar,
	}
}

type Buzzable interface {
	DoBuzz()
}

type Buzz struct {
	Bazz Bazzable
}

func (b Buzz) DoBuzz() {
	fmt.Println("buzz")
	b.Bazz.DoBazz()
}

func NewBuzz(bazz Bazzable) Buzzable {
	return &Buzz{
		Bazz: bazz,
	}
}

type Burrable interface {
	DoBurr()
}

type Burr struct {
	Buzz Buzzable
	Bazz *Bazz
}

func (b Burr) DoBurr() {
	fmt.Println("burr")
	b.Buzz.DoBuzz()
	b.Bazz.DoBazz()
}

func NewBurr(buzz Buzzable, bazz *Bazz) Burrable {
	return &Burr{
		Buzz: buzz,
		Bazz: bazz,
	}
}

func TestRegisterSingleton(t *testing.T) {
	c := NewContainer()

	RegisterSingleton[*Foo](c, NewFoo)
	RegisterSingleton[*Bar](c, NewBar)
	RegisterSingleton[Bazzable](c, NewBazz)
	RegisterSingleton[Buzzable](c, NewBuzz)
	RegisterSingleton[Burrable](c, NewBurr)

	buzz := Get[Buzzable](c)
	buzz.DoBuzz()

	burr := Get[Burrable](c)
	burr.DoBurr()
}

func TestGenerateDependencyGraph(t *testing.T) {
	c := NewContainer()

	RegisterSingleton[*Foo](c, NewFoo)
	RegisterSingleton[*Bar](c, NewBar)
	RegisterSingleton[Bazzable](c, NewBazz)
	RegisterSingleton[Buzzable](c, NewBuzz)
	RegisterSingleton[Burrable](c, NewBurr)

	depGraph := GenerateDependencyGraph(c)
	fmt.Println(depGraph)
}

type Barkable interface {
	DoBark()
}

type Dog struct {
}

func (d Dog) DoBark() {
	fmt.Println("woof")
}

var counter = 0

func NewDog() Barkable {
	counter++
	return &Dog{}
}

func TestRegisterGenerator(t *testing.T) {
	c := NewContainer()

	RegisterGenerator[Barkable](c, NewDog)

	dog1 := Get[Barkable](c)
	dog1.DoBark()

	dog2 := Get[Barkable](c)
	dog2.DoBark()

	if counter != 2 {
		t.Errorf("expected 2 dogs, got %d", counter)
	}
}

type ToFill struct {
	Foo      *Foo
	Bar      *Bar
	Bazzable Bazzable
	Burrable Burrable `inject:"ignore"`
}

func (t ToFill) String() string {
	return fmt.Sprintf("%v %v %v %v", t.Foo, t.Bar, t.Bazzable, t.Burrable)
}

func TestForStruct(t *testing.T) {
	c := NewContainer()

	RegisterSingleton[*Foo](c, NewFoo)
	RegisterSingleton[*Bar](c, NewBar)
	RegisterSingleton[Bazzable](c, NewBazz)
	RegisterSingleton[Buzzable](c, NewBuzz)
	RegisterSingleton[Burrable](c, NewBurr)

	toFill := ForStruct[ToFill](c)
	fmt.Println(toFill.String())

	// check that the ignored field is nil
	if toFill.Burrable != nil {
		t.Errorf("expected nil, got %v", toFill.Burrable)
	}

	// check that the other fields are not nil
	if toFill.Foo == nil {
		t.Errorf("expected non-nil, got %v", toFill.Foo)
	}

	if toFill.Bar == nil {
		t.Errorf("expected non-nil, got %v", toFill.Bar)
	}

	if toFill.Bazzable == nil {
		t.Errorf("expected non-nil, got %v", toFill.Bazzable)
	}
}

func TestCreatorIsFunction(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	RegisterSingleton[*Foo](c, 1)
}

func TestCreatorReturnsSingleValue(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	RegisterSingleton[*Foo](c, func() (*Foo, *Foo) {
		return &Foo{}, &Foo{}
	})
}

func TestCreatorReturnsPointer(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	RegisterSingleton[*Foo](c, func() Foo {
		return Foo{}
	})
}

func TestCreatorReturnsStructThatImplementsInterface(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic")
		}
	}()

	RegisterSingleton[Bazzable](c, func() *Bazz {
		return &Bazz{}
	})
}

func TestCreatorReturnsStructThatDoesNotImplementInterface(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	RegisterSingleton[Bazzable](c, func() *Foo {
		return &Foo{}
	})
}

func TestCreatorReturnsNil(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	RegisterSingleton[Bazzable](c, func() Bazzable {
		return nil
	})
}

func TestTypeParameterIsEitherPointerOrInterface(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	RegisterSingleton[Bar](c, NewBar)
}

func TestCreatorDoesNotReturnPointerToInterface(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	RegisterSingleton[Bazzable](c, func() *Bazzable {
		return nil
	})
}

func TestCreatorReturnsPointerToStruct(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	RegisterSingleton[*Foo](c, func() Foo {
		return Foo{}
	})
}

func TestGetForGeneratorDoesNotReturnNil(t *testing.T) {
	c := NewContainer()

	RegisterGenerator[Barkable](c, func() *Dog {
		return nil
	})

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	Get[Barkable](c)
}

func TestTypeIsRegistered(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	Get[Barkable](c)
}

func TestTypeForCreatorParameterIsRegistered(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	RegisterSingleton[Barkable](c, func(baz *Bazz) *Dog {
		return &Dog{}
	})
}

func TestInjectFieldIsNotRegistered(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	RegisterSingleton[*Foo](c, NewFoo)
	RegisterSingleton[*Bar](c, NewBar)

	ForStruct[ToFill](c)
}

func TestForStructIsActuallyAStruct(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	ForStruct[int](c)
}

func TestForFuncIsActuallyAFunction(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	ForFunc(c, 1)
}

func TestForFuncFunctionDoesNotReturnAnything(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	ForFunc(c, func() uint8 { return 1 })
}

func TestForFunc(t *testing.T) {
	c := NewContainer()

	RegisterSingleton[*Foo](c, NewFoo)
	RegisterSingleton[*Bar](c, NewBar)

	ForFunc(c, func(foo *Foo, bar *Bar) {
		if foo == nil {
			t.Errorf("foo is nil")
		}

		if bar == nil {
			t.Errorf("bar is nil")
		}
	})
}

func TestForFuncHasUnregisteredParameter(t *testing.T) {
	c := NewContainer()

	// this should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	ForFunc(c, func(foo *Foo) {})
}

func TestForFuncAsync(t *testing.T) {
	c := NewContainer()

	RegisterSingleton[*Foo](c, NewFoo)
	RegisterSingleton[*Bar](c, NewBar)

	doneChan := make(chan struct{})

	ForFuncAsync(c, func(foo *Foo, bar *Bar) {
		if foo == nil {
			t.Errorf("foo is nil")
		}

		if bar == nil {
			t.Errorf("bar is nil")
		}

		fmt.Println("done")

		doneChan <- struct{}{}
	})

	fmt.Println("waiting for done")

	<-doneChan
}
