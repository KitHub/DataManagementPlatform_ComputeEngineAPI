package entity

import (
	"fmt"
	"strings"
)

type SetOperator int32
type SetOperatorNodeType int32

const (
	SetOperatorNodeTypeUnspecified SetOperatorNodeType = 0
	SetOperatorNodeTypeDataSet     SetOperatorNodeType = 1
	SetOperatorNodeTypeOperator    SetOperatorNodeType = 2
)

const (
	SetOperatorUnspecified SetOperator = 0
	SetOperatorUnion       SetOperator = 1
	SetOperatorIntersect   SetOperator = 2
	SetOperatorDiff        SetOperator = 3
)

var SetOperatorSymbol = map[SetOperator]string{
	SetOperatorUnion:     "∪",
	SetOperatorIntersect: "∩",
	SetOperatorDiff:      "-",
}

type SetOperationNode struct {
	NodeType    SetOperatorNodeType
	SetOperator SetOperator
	Data        string
	Children    []*SetOperationNode
}

func (n *SetOperationNode) String() string {
	return n.printNode(SetOperatorUnspecified)
}

func (n *SetOperationNode) printNode(parentOp SetOperator) string {
	switch n.NodeType {
	case SetOperatorNodeTypeDataSet:
		return n.Data

	case SetOperatorNodeTypeOperator:
		childStrs := make([]string, 0, len(n.Children))
		for _, child := range n.Children {
			childStrs = append(childStrs, child.printNode(n.SetOperator))
		}

		var expr string
		switch n.SetOperator {
		case SetOperatorUnion, SetOperatorIntersect:
			expr = strings.Join(childStrs, fmt.Sprintf(" %s ", SetOperatorSymbol[n.SetOperator]))
		case SetOperatorDiff:
			expr = fmt.Sprintf("%s %s %s", childStrs[0], SetOperatorSymbol[SetOperatorDiff], childStrs[1])
		default:
			expr = "unknown_operator: " + fmt.Sprintf("%d", n.SetOperator)

		}

		if parentOp != SetOperatorUnspecified && parentOp != n.SetOperator {
			return fmt.Sprintf("(%s)", expr)
		}
		return expr
	default:
		return "unknown_node_type: " + fmt.Sprintf("%d", n.NodeType)
	}
}
