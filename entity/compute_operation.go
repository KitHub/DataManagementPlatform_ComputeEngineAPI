package entity

type SetOperator int32

const (
	SetOperatorUnspecified SetOperator = 0
	SetOperatorUnion       SetOperator = 1
	SetOperatorIntersect   SetOperator = 2
	SetOperatorDiff        SetOperator = 3
)

type SetOperationNode struct {
	SetOperator SetOperator
	Data        string
	LeftNode    *SetOperationNode
	RightNode   *SetOperationNode
}
