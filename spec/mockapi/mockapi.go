package mockapi

import (
	"net/http/httputil"
)

func Dial(param1 string) *Obj {
	return nil
}

// Comment for PublicEmbedded
type PublicEmbedded struct {
}

// Comment for FuncInPublicEmbedded
func (*PublicEmbedded) FuncInPublicEmbedded() {}

type privateEmbedded struct {
}

// Comment for FuncInPrivateEmbedded
func (*privateEmbedded) FuncInPrivateEmbedded() {}

//Comment for signature
type Signature func(param1 string, param2 int, param3 *int, param4 *SubObj, param5 SubObj, paramVar6 Interface1) string

//Comment for Obj
type Obj struct {
	*PublicEmbedded
	*privateEmbedded
	//Comment for FieldFunc
	FieldFunc1 func(

		func2 Signature,
		var1 string,
		var2 int,
		var3 *int,
		var4 *SubObj,
		var5 SubObj,
		var6 Interface1,
		slice1 []*SubObj,
		slice2 []SubObj,
		slice3 []int,
		slice4 NamedSlice,
		chan1chan interface{},
		chan2 chan struct{},
		chan3 chan struct {
			A string
			B string
		},
		chan4 chan *struct {
			A string
			B string
		},
		chan5 chan *SubObj,
		chan6 chan SubObj,
		depObj *DepObj,

	) (string, *Obj, httputil.BufferPool, error)
	Func2 Signature
	Var1  string
	Var2  int
	Var3  *int
	Var4  *SubObj
	Var5  SubObj
	//Comment for variable
	Var6   Interface1
	Slice1 []*SubObj
	Slice2 []SubObj
	Slice3 []int
	Slice4 NamedSlice
	Chan1  interface{}
	Chan2  chan struct{}
	Chan3  chan struct {
		A string
		B string
	}
	Chan4 chan *struct {
		A string
		B string
	}
	Chan5  chan *SubObj
	Chan6  chan SubObj
	DepObj *DepObj
}

//Comment for NamedSlice
type NamedSlice []*SubObj

func (obj Obj) MemberNonPointerReceiver() {

}

type SubObj struct {
}

//Deprecated: This will be removed in a future version
type DepObj struct{}

// Stability: Experimental
type ExperimentalObject1 struct{}

// Experimental comment
// Stability: Experimental
type ExperimentalObject2 struct{}

// Experimental comment
// Stability: Experimental
func (ExperimentalObject1) ExpMember1() {

}

//Deprecated: This will be removed in a future version
func (*DepObj) DepMember1() {

}

//Comment preceding deprecation
//Deprecated: This will be removed in a future version
func (*DepObj) DepMember2() {

}

// Comment above type block that shouldn't appear anywhere in the spec
type (
	// Comment PublicEmbeddedInteface
	PublicEmbeddedInteface interface {
		FuncInPublicEmbeddedInterface()
	}
	// Comment privateEmbeddedInteface
	privateEmbeddedInteface interface {
		FuncInPrivateEmbeddedInterface()
	}
	// Comment Interface1
	Interface1 interface {
		PublicEmbeddedInteface
		privateEmbeddedInteface
		InterfaceFunc1(param string) string
	}
)

// Comment for Member
func (obj *Obj) Member(
	var1 string,
	var2 int,
	var3 *int,
	var4 *SubObj,
	var5 SubObj,
	var6 Interface1,
	slice1 []*SubObj,
	slice2 []SubObj) {

}

// Alias for float32 so we can add some methods
type Float32 float32

//Method on Float32
func (float32 *Float32) AdditionalMethodOnFloat() {}
