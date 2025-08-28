package nodeStore

type LogicManager struct {
	self  []byte
	logic [][]byte
}

func NewLogicManager(id []byte) *LogicManager {
	lm := LogicManager{
		self:  id,
		logic: make([][]byte, 0),
	}
	return &lm
}
