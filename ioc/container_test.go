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

func Cleanup() {
	implMapMu.Lock()
	defer implMapMu.Unlock()
	implMap = make(map[string]implementationDetails)
}

func TestRegisterImpl(t *testing.T) {
	RegisterImpl[*Foo](NewFoo)
	RegisterImpl[*Bar](NewBar)
	RegisterImpl[Bazzable](NewBazz)
	RegisterImpl[Buzzable](NewBuzz)
	RegisterImpl[Burrable](NewBurr)

	buzz := GetImpl[Buzzable]()
	buzz.DoBuzz()

	burr := GetImpl[Burrable]()
	burr.DoBurr()

	Cleanup()
}

func TestGenerateDependencyGraph(t *testing.T) {
	RegisterImpl[*Foo](NewFoo)
	RegisterImpl[*Bar](NewBar)
	RegisterImpl[Bazzable](NewBazz)
	RegisterImpl[Buzzable](NewBuzz)
	RegisterImpl[Burrable](NewBurr)

	depGraph := GenerateDependencyGraph()
	fmt.Println(depGraph)

	Cleanup()
}
