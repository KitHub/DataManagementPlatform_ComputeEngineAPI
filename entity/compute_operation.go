package entity

type SetOperator int32

const (
	SetOperatorAnd  SetOperator = 0
	SetOperatorOr   SetOperator = 1
	SetOperatorDiff SetOperator = 2
)

type SetOperationNode struct {
	SetOperator SetOperator
	LeftNode    *SetOperationNode
	RightNode   *SetOperationNode
}
