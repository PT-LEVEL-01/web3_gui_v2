package leveldb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sort"
)

// 1个key对应多个field（用于批量查询）
type KFPair struct {
	Key    []byte
	Fields [][]byte
}

// 一个field：value对
type FVPair struct {
	Field []byte
	Value []byte
}

// 一个key对应多个field：value对（用于批量更新）
type KFVPair struct {
	Key    []byte
	Values []FVPair
}

var errHashKey = errors.New("invalid hash key")

const (
	hashStartSep byte = ':'
	hashStopSep  byte = hashStartSep + 1
)

// 验证key和field的大小
func checkHashKFSize(key []byte, field []byte) error {
	if len(key) > MaxKeySize || len(key) == 0 {
		return errKeySize
	} else if len(field) > MaxHashFieldSize || len(field) == 0 {
		return errHashFieldSize
	}
	return nil
}

// 编码key的大小
func (l *LevelDB) hEncodeSizeKey(key []byte) []byte {
	buf := make([]byte, len(key)+1+len(l.indexVarBuf))

	pos := copy(buf, l.indexVarBuf)

	buf[pos] = HSizeType

	pos++
	copy(buf[pos:], key)

	return buf
}

// 得到前缀为key起始键
func (l *LevelDB) hEncodeStartKey(key []byte) []byte {
	return l.hEncodeHashKey(key, nil)
}

// 得到前缀为key结束键
func (l *LevelDB) hEncodeStopKey(key []byte) []byte {
	k := l.hEncodeHashKey(key, nil)

	k[len(k)-1] = hashStopSep

	return k
}

// 编码key和field合并之后的键
func (l *LevelDB) hEncodeHashKey(key []byte, field []byte) []byte {
	buf := make([]byte, len(key)+len(field)+1+1+2+len(l.indexVarBuf))

	pos := copy(buf, l.indexVarBuf)

	buf[pos] = HashType
	pos++

	binary.BigEndian.PutUint16(buf[pos:], uint16(len(key)))
	pos += 2

	copy(buf[pos:], key)
	pos += len(key)

	buf[pos] = hashStartSep
	pos++

	copy(buf[pos:], field)

	return buf
}

// 解码key、field合并的键
func (l *LevelDB) hDecodeHashKey(ek []byte) ([]byte, []byte, error) {
	pos, err := l.checkKeyIndex(ek)
	if err != nil {
		return nil, nil, err
	}

	if pos+1 > len(ek) || ek[pos] != HashType {
		return nil, nil, errHashKey
	}
	pos++

	if pos+2 > len(ek) {
		return nil, nil, errHashKey
	}

	keyLen := int(binary.BigEndian.Uint16(ek[pos:]))
	pos += 2

	if keyLen+pos > len(ek) {
		return nil, nil, errHashKey
	}

	key := ek[pos : pos+keyLen]
	pos += keyLen
	if ek[pos] != hashStartSep {
		return nil, nil, errHashKey
	}
	pos++
	field := ek[pos:]
	return key, field, nil
}

// hash 设置单个key和field的值
func (l *LevelDB) HSet(key, field, value []byte) (int64, error) {
	if err := checkHashKFSize(key, field); err != nil {
		return 0, err
	} else if err := checkValueSize(value); err != nil {
		return 0, err
	}

	//snap
	ek := l.hEncodeHashKey(key, field)
	return 1, l.snap.Put(ek, value)
}

// hash查询key下的field
func (l *LevelDB) HGet(key, field []byte) ([]byte, error) {
	if err := checkHashKFSize(key, field); err != nil {
		return nil, err
	}
	//snap
	return l.snap.Get(l.hEncodeHashKey(key, field))
}

// hash查询key下的多个field
func (l *LevelDB) HMget(key []byte, fields ...[]byte) ([][]byte, error) {
	//snap
	ks := make([][]byte, len(fields))
	for i := 0; i < len(fields); i++ {
		if err := checkHashKFSize(key, fields[i]); err != nil {
			return nil, err
		}

		ks[i] = l.hEncodeHashKey(key, fields[i])
	}
	return l.snap.MGet(ks...)
}

// hash设置key下的多个field
func (l *LevelDB) HMset(key []byte, args ...FVPair) error {
	//snap
	for i := 0; i < len(args); i++ {
		if err := checkHashKFSize(key, args[i].Field); err != nil {
			return err
		} else if err := checkValueSize(args[i].Value); err != nil {
			return err
		}

		l.snap.Put(l.hEncodeHashKey(key, args[i].Field), args[i].Value)
	}

	return nil
}

// hash设置多个key下的多个field
func (l *LevelDB) HKMset(kfvs []KFVPair) error {
	//snap
	for i := 0; i < len(kfvs); i++ {
		for j := 0; j < len(kfvs[i].Values); j++ {
			ek := l.hEncodeHashKey(kfvs[i].Key, kfvs[i].Values[j].Field)
			l.snap.Put(ek, kfvs[i].Values[j].Value)
		}
	}

	return nil
}

// hash 获取多个key下的多个field
func (l *LevelDB) HKMget(kfs []KFPair) ([][][]byte, error) {
	//snap
	r := make([][][]byte, len(kfs))
	for i := 0; i < len(kfs); i++ {
		ks1 := make([][]byte, len(kfs[i].Fields))
		for j := 0; j < len(kfs[i].Fields); j++ {
			ek := l.hEncodeHashKey(kfs[i].Key, kfs[i].Fields[j])
			ks1[j] = ek
		}
		r1, _ := l.snap.MGet(ks1...)
		r[i] = r1
	}

	return r, nil
}

// hash查询key下的所有field:value
func (l *LevelDB) HGetAll(key []byte) ([]FVPair, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}

	start := l.hEncodeStartKey(key)
	stop := l.hEncodeStopKey(key)

	//此时kvs包含了db和增量中的所有范围内的key，结果为无序的
	kvs := l.snap.GetInRange(start, stop, RangeROpen)

	//排序
	nks := make([][]byte, len(kvs))
	i := 0
	for k := range kvs {
		nks[i] = l.snap.decodeKey(k)
		i++
	}
	sort.SliceStable(nks, func(i, j int) bool {
		return bytes.Compare(nks[i], nks[j]) < 0
	})

	pairs := make([]FVPair, len(kvs))
	for i := range nks {
		//解析key和field
		_, f, err := l.hDecodeHashKey(nks[i])
		if err != nil {
			return nil, err
		}

		pairs[i] = FVPair{f, kvs[l.snap.encodeKey(nks[i])]}
	}

	return pairs, nil
}

func (l *LevelDB) hSetItem(key, field, value []byte) int64 {
	b := l.hashBatch

	ek := l.hEncodeHashKey(key, field)

	var n int64 = 1
	if v, _ := l.Get(ek); v != nil {
		n = 0
	}

	b.Put(ek, value)
	return n
}

// 根据key删除多个field
func (l *LevelDB) HDel(key []byte, args ...[]byte) (int64, error) {
	//snap
	ks := make([][]byte, len(args))
	for i := 0; i < len(args); i++ {
		if err := checkHashKFSize(key, args[i]); err != nil {
			return 0, err
		}

		ks[i] = l.hEncodeHashKey(key, args[i])
	}

	l.snap.Del(ks...)

	return 0, nil
}

func (l *LevelDB) hDelete(t *Batch, key []byte) int64 {
	sk := l.hEncodeSizeKey(key)
	start := l.hEncodeStartKey(key)
	stop := l.hEncodeStopKey(key)

	var num int64
	it := l.RangeLimitIterator(start, stop, RangeROpen, 0, -1)
	for ; it.Valid(); it.Next() {
		t.Delete(it.it.Key())
		num++
	}

	t.Delete(sk)
	return num
}

// 清除指定key的所有存储
func (l *LevelDB) HClear(key []byte) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return 0, err
	}

	//snap
	start := l.hEncodeStartKey(key)
	stop := l.hEncodeStopKey(key)
	l.snap.DelInRange(start, stop)

	return 0, nil
}

// 解码key、field合并的键
func (l *LevelDB) HDecodeHashKey(ek []byte) ([]byte, []byte, error) {
	return l.hDecodeHashKey(ek)
}

// 编码key、field合并的键
func (l *LevelDB) HEncodeHashKey(key, field []byte) []byte {
	return l.hEncodeHashKey(key, field)
}
