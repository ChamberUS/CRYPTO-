package types

import proto "github.com/cosmos/gogoproto/proto"

type Module struct {
	Authority string `protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty" yaml:"authority"`
}

var _ proto.Message = (*Module)(nil)

// Reset implements proto.Message.
func (m *Module) Reset() { *m = Module{} }

// String implements proto.Message.
func (m *Module) String() string { return proto.CompactTextString(m) }

// ProtoMessage implements proto.Message.
func (*Module) ProtoMessage() {}

// Descriptor is required by the legacy proto.Message interface.
func (*Module) Descriptor() ([]byte, []int) { return []byte{}, []int{0} }

func init() {
	// Register the type URL so appconfig.WrapAny can pack/unpack this config.
	proto.RegisterType((*Module)(nil), "byx.feesplit.module.v1.Module")
}
