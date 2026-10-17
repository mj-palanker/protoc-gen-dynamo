// Package dynamoattr reads and writes the item-level attributes that
// protoc-gen-dynamo emits outside the marshalled message.
package dynamoattr

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"google.golang.org/protobuf/proto"

	dynamov1 "github.com/pquerna/protoc-gen-dynamo/dynamo/v1"
)

// WritePriorityAttr is the DynamoDB attribute name holding the write priority.
//
// The priority describes how the change carrying it should be processed, not
// the object: it is stamped on the item only so change-stream consumers can
// see it on the stream record.
const WritePriorityAttr = "wp"

var (
	priorityToAttr = map[dynamov1.WritePriority]string{}
	attrToPriority = map[string]dynamov1.WritePriority{}
)

// The maps are built from the attr_value enum-value option, the same source
// the generator reads, so the two cannot disagree.
func init() {
	values := dynamov1.WritePriority(0).Descriptor().Values()
	for i := 0; i < values.Len(); i++ {
		ev := values.Get(i)
		s, _ := proto.GetExtension(ev.Options(), dynamov1.E_AttrValue).(string)
		if s == "" {
			continue
		}
		p := dynamov1.WritePriority(ev.Number())
		priorityToAttr[p] = s
		attrToPriority[s] = p
	}
	priorityToAttr[dynamov1.WritePriority_WRITE_PRIORITY_UNSPECIFIED] = priorityToAttr[dynamov1.WritePriority_WRITE_PRIORITY_FAST]
}

// WritePriorityAttrValue returns the stored string for p. UNSPECIFIED maps to
// the FAST value. A value unknown to this build (e.g. a number from a newer
// producer) also maps to FAST.
func WritePriorityAttrValue(p dynamov1.WritePriority) string {
	if s, ok := priorityToAttr[p]; ok {
		return s
	}
	return priorityToAttr[dynamov1.WritePriority_WRITE_PRIORITY_FAST]
}

// ParseWritePriority maps a stored string back to the enum. ok is false for an
// empty or unrecognised value (e.g. one written by a newer producer); callers
// must handle that rather than fail.
func ParseWritePriority(s string) (dynamov1.WritePriority, bool) {
	p, ok := attrToPriority[s]
	return p, ok
}

// SetWritePriority overwrites the attribute on a marshalled item, e.g. the
// result of MarshalDynamoItem or a PutItemInput.Item.
func SetWritePriority(item map[string]types.AttributeValue, p dynamov1.WritePriority) {
	item[WritePriorityAttr] = &types.AttributeValueMemberS{Value: WritePriorityAttrValue(p)}
}

// GetWritePriority reads the attribute from a marshalled item. ok is false if
// the attribute is missing (e.g. an item written before it existed), is not a
// string, or holds an unrecognised value.
func GetWritePriority(item map[string]types.AttributeValue) (dynamov1.WritePriority, bool) {
	av, ok := item[WritePriorityAttr].(*types.AttributeValueMemberS)
	if !ok {
		return dynamov1.WritePriority_WRITE_PRIORITY_UNSPECIFIED, false
	}
	return ParseWritePriority(av.Value)
}
