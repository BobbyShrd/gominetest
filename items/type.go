package items

import (
	"fmt"
	"github.com/irmine/gonbt"
	"strings"
)

// Type is the type that identifies an item.
// Types contain a string ID,
// which can be used to construct a new item stack.
type Type struct {
	// NBTParseFunction gets called once NBT is attempted
	// to be decoded for an item. The compound passed is the
	// compound the NBT data should be coming out of, and the stack
	// passed is the stack that encapsulates this type.
	NBTParseFunction func(compound *gonbt.Compound, stack *Stack)
	// NBTEmitFunction gets called once NBT is attempted
	// to be obtained from an item. The compound passed is the
	// compound the NBT data should be going into, and the stack
	// passed is the stack that encapsulates this type.
	NBTEmitFunction func(compound *gonbt.Compound, stack *Stack)

	// name is the name of the item type.
	// This name is merely a modification of the string ID.
	name string
	// stringId is the identifier of the item type.
	// This string ID is always used, rather than numeric IDs.
	stringId string
	// breakable defines if the item is breakable.
	// Breakable items will have decrementing durability.
	breakable bool
	// maxStackSize is the maximum size of a stack of this item.
	// Item stacks itself are not limited, but the stack size
	// of occurrences in an inventory of the item are.
	maxStackSize int
}

// NewType returns a new non-breakable type.
// The given string ID is used as identifier,
// and all properties are immune in the type.
// Type names prefixed with `minecraft:` get their
// name set to it without the prefix.
// Types get the default NBT parsing and emitting functions.
func NewType(stringId string) Type {
	fragments := strings.Split(stringId[10:], "_")
	name := ""
	for _, frag := range fragments {
		name += strings.Title(frag) + " "
	}
	return Type{ParseNBT, EmitNBT, strings.TrimRight(name, " "), stringId, false, 64}
}

// NewType returns a new breakable type.
// The given string ID is used as identifier,
// and all properties are immune in the type.
// Type names prefixed with `minecraft:` get their
// name set to it without the prefix.
// Types get the default NBT parsing and emitting functions.
func NewBreakable(stringId string) Type {
	t := NewType(stringId)
	t.breakable = true
	return t
}

// GetName returns the readable name of an item type.
// This name may contains spaces.
func (t Type) GetName() string {
	return t.name
}

// GetId returns the string ID of an item type.
// StringIds are a string used as an identifier,
// in order to lookup items by it.
func (t Type) GetId() string {
	return t.stringId
}

// IsBreakable checks if an item is breakable.
// Breakable items use data fields for durability,
// but we separate them for forward compatibility sake.
func (t Type) IsBreakable() bool {
	return t.breakable
}

// GetMaximumStackSize returns the maximum stack size of an item.
// Item stacks of the type are not limited to this size themselves,
// but are when set into an inventory.
func (t Type) GetMaximumStackSize() int {
	return t.maxStackSize
}

// String returns a string representation of a type.
// It implements fmt.Stringer, and returns a string as such:
// Emerald(minecraft:emerald)
func (t Type) String() string {
	return fmt.Sprint(t.name, "(", t.stringId, ")")
}

// GetAuxValue returns the aux value for the item stack with item data.
// This aux value is used for writing stacks over network.
func (t Type) GetAuxValue(stack *Stack, data int16) int32 {
	if t.IsBreakable() {
		data = stack.Durability
	}
	return int32(((data & 0x7fff) << 8) | int16(stack.Count))
}

// Equals checks if two item types are considered equal.
// Item types are merely checked against each other's
// string IDs, but should not require more comparisons.
func (t Type) Equals(t2 Type) bool {
	return t.stringId == t2.stringId
}

// ParseNBT implements default behaviour for parsing NBT.
// This is the default function passed in for `NBTParseFunction`.
// The cached NBT gets set when parsing NBT.
func ParseNBT(compound *gonbt.Compound, stack *Stack) {
	if compound.HasTagWithType(Display, gonbt.TAG_Compound) {
		stack.DisplayName = compound.GetCompound(Display).GetString(DisplayName, stack.name)
		for _, tag := range compound.GetCompound(Display).GetList(DisplayLore, gonbt.TAG_String).GetTags() {
			stack.Lore = append(stack.Lore, tag.Interface().(string))
		}
	}
	stack.cachedNBT = compound
}

// EmitNBT implements default behaviour for emitting NBT.
// This is the default function passed in for `NBTEmitFunction`.
// The compound first gets set to the cached compound of the item type.
func EmitNBT(compound *gonbt.Compound, stack *Stack) {
	compound = stack.cachedNBT
	compound.SetCompound(Display, make(map[string]gonbt.INamedTag))
	if stack.DisplayName != "" {
		compound.GetCompound(Display).SetString(DisplayName, stack.DisplayName)
		var list []gonbt.INamedTag
		for _, lore := range stack.Lore {
			list = append(list, gonbt.NewString("", lore))
		}
		compound.GetCompound(Display).SetList(DisplayLore, gonbt.TAG_String, list)
	}
}
