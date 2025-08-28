package leveldb

import (
	"bytes"

	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	IteratorForward  uint8 = 0
	IteratorBackward uint8 = 1
)
const (
	RangeClose uint8 = 0x00
	RangeLOpen uint8 = 0x01
	RangeROpen uint8 = 0x10
	RangeOpen  uint8 = 0x11
)

type Iterator struct {
	iterator.Iterator
}

//新建一个迭代器
func NewIterator(l *LevelDB) *Iterator {
	return &Iterator{l.db.NewIterator(nil, nil)}
}

//新建一个带选项的迭代器
func NewIteratorWithOption(l *LevelDB, slice *util.Range, ro *opt.ReadOptions) *Iterator {
	return &Iterator{l.db.NewIterator(slice, ro)}
}

//迭代器查询(批量查询用)
func (it *Iterator) Find(key []byte) []byte {
	it.Seek(key)
	if it.Valid() {
		k := it.Key()
		if k == nil {
			return nil
		} else if bytes.Equal(k, key) {
			return it.Value()
		}
	}

	return nil
}

//迭代器范围选项
//range type:
//close: [min, max]
//open: (min, max)
//lopen: (min, max]
//ropen: [min, max)
type Range struct {
	Min []byte //起始key
	Max []byte //结束key

	Type uint8 //查询类型，0x01为往左查询且不包含最小值，0x10为往右查询且不包含最大值
}

type Limit struct {
	Offset int //范围查询偏移量
	Count  int //查询的数量
}

type RangeLimitIterator struct {
	it *Iterator

	r *Range
	l *Limit

	step int

	//0 for IteratorForward, 1 for IteratorBackward
	direction uint8
}

//校验查询的key
func (it *RangeLimitIterator) Valid() bool {
	if it.l.Offset < 0 {
		return false
	} else if !it.it.Valid() {
		return false
	} else if it.l.Count >= 0 && it.step >= it.l.Count {
		return false
	}

	if it.direction == IteratorForward {
		if it.r.Max != nil {
			r := bytes.Compare(it.it.Key(), it.r.Max)
			if it.r.Type&RangeROpen > 0 {
				return !(r >= 0)
			} else {
				return !(r > 0)
			}
		}
	} else {
		if it.r.Min != nil {
			r := bytes.Compare(it.it.Key(), it.r.Min)
			if it.r.Type&RangeLOpen > 0 {
				return !(r <= 0)
			} else {
				return !(r < 0)
			}
		}
	}

	return true
}

//迭代下一个key
func (it *RangeLimitIterator) Next() {
	it.step++

	if it.direction == IteratorForward {
		it.it.Next()
	} else {
		it.it.Prev()
	}
}

func NewRangeLimitIterator(i *Iterator, r *Range, l *Limit) *RangeLimitIterator {
	return rangeLimitIterator(i, r, l, IteratorForward)
}

//迭代器范围查询
func rangeLimitIterator(i *Iterator, r *Range, l *Limit, direction uint8) *RangeLimitIterator {
	it := new(RangeLimitIterator)

	it.it = i

	it.r = r
	it.l = l
	it.direction = direction

	it.step = 0

	if l.Offset < 0 {
		return it
	}

	if direction == IteratorForward {
		if r.Min == nil {
			it.it.First()
		} else {
			it.it.Seek(r.Min)

			if r.Type&RangeLOpen > 0 {
				if it.it.Valid() && bytes.Equal(it.it.Key(), r.Min) {
					it.it.Next()
				}
			}
		}
	} else {
		if r.Max == nil {
			it.it.Last()
		} else {
			it.it.Seek(r.Max)

			if !it.it.Valid() {
				it.it.Last()
			} else {
				if !bytes.Equal(it.it.Key(), r.Max) {
					it.it.Prev()
				}
			}

			if r.Type&RangeROpen > 0 {
				if it.it.Valid() && bytes.Equal(it.it.Key(), r.Max) {
					it.it.Prev()
				}
			}
		}
	}

	for i := 0; i < l.Offset; i++ {
		if it.it.Valid() {
			if it.direction == IteratorForward {
				it.it.Next()
			} else {
				it.it.Prev()
			}
		}
	}

	return it
}

func (it *RangeLimitIterator) Iterator() *Iterator {
	return it.it
}
