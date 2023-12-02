package stream

import (
	"github.com/fatih/structs"
	"testing"
)

type Child struct {
	ChildA string
}

type Struct struct {
	Chd *Child
}

func TestStructField(t *testing.T) {
	s := &Struct{
		Chd: &Child{
			ChildA: "hello world",
		},
	}

	obj := structs.New(s)
	val := obj.Field("Chd").Field("ChildA")
	if val.Value().(string) != "hello world" {
		t.Fatal("unsupported")
	}
}

func TestNonExistentField(t *testing.T) {
	s := &Struct{
		Chd: &Child{
			ChildA: "hello world",
		},
	}

	obj := structs.New(s)
	//val := obj.Field("Chd").Field("ChildB")//panic
	if field, ok := obj.FieldOk("Chd"); ok {
		if fieldB, ok := field.FieldOk("ChildB"); ok {
			println("fieldB", fieldB.Value().(string))
		}
	}
}
