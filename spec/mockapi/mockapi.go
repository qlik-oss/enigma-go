package mockapi

func Dial(param1 string) *Obj {
	return nil
}

type PublicEmbedded struct {
}

func (*PublicEmbedded) FuncInPublicEmbedded() {}

type privateEmbedded struct {
}

func (*privateEmbedded) FuncInPrivateEmbedded() {}

type Signature func(param1 string, param2 int, param3 *int, param4 *SubObj, param5 SubObj, paramVar6 Interface1) string

type Obj struct {
	*PublicEmbedded
	*privateEmbedded
	FieldFunc1 func(param1 string, param2 int, param3 *int, param4 *SubObj, param5 SubObj, paramVar6 Interface1) string
	FieldFunc2 Signature
	FieldVar1  string
	FieldVar2  int
	FieldVar3  *int
	FieldVar4  *SubObj
	FieldVar5  SubObj
	FieldVar6  Interface1
}

type SubObj struct {
}

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

func (obj *Obj) member(param1 string) {

}
