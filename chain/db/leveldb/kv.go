package leveldb

import "errors"

//kv对，批量操作用
type KVPair struct {
	Key   []byte
	Value []byte
}

//验证key大小
func checkKeySize(key []byte) error {
	if len(key) > MaxKeySize || len(key) == 0 {
		return errKeySize
	}
	return nil
}

//验证value大小
func checkValueSize(value []byte) error {
	if len(value) > MaxValueSize {
		return errValueSize
	}

	return nil
}

//解码key
func (l *LevelDB) DecodeKVKey(ek []byte) ([]byte, error) {
	pos, err := l.checkKeyIndex(ek)
	if err != nil {
		return nil, err
	}
	if pos+1 > len(ek) || ek[pos] != KVType {
		return nil, errors.New("errKVKey")
	}

	pos++

	return ek[pos:], nil
}

//编码key
func (l *LevelDB) encodeKVKey(key []byte) []byte {
	ek := make([]byte, len(key)+1+len(l.indexVarBuf))

	pos := copy(ek, l.indexVarBuf)
	ek[pos] = KVType
	pos++

	copy(ek[pos:], key)
	return ek
}

//获取单个k的值
func (l *LevelDB) Get(key []byte) ([]byte, error) {
	if err := checkKeySize(key); err != nil {
		return nil, err
	}

	key = l.encodeKVKey(key)

	//snap
	v, err := l.snap.Get(key)

	if err != nil && err.Error() == ErrNotFound.Error() {
		return nil, nil
	}
	return v, nil
}

//设置单个kv
func (l *LevelDB) Set(key, value []byte) error {
	if err := checkKeySize(key); err != nil {
		return err
	} else if err := checkValueSize(value); err != nil {
		return err
	}

	key = l.encodeKVKey(key)

	//snap
	return l.snap.Put(key, value)
}

//批量获取多个k的值
func (l *LevelDB) MGet(keys ...[]byte) ([][]byte, error) {
	//snap
	ks := make([][]byte, len(keys))
	for i := range keys {
		if err := checkKeySize(keys[i]); err != nil {
			return nil, err
		}
		ks[i] = l.encodeKVKey(keys[i])
	}

	return l.snap.MGet(ks...)
}

//批量设置多个kv对
func (l *LevelDB) MSet(kvs ...KVPair) error {
	//snap
	for i := 0; i < len(kvs); i++ {
		if err := checkKeySize(kvs[i].Key); err != nil {
			return err
		} else if err := checkValueSize(kvs[i].Value); err != nil {
			return err
		}

		l.snap.Put(l.encodeKVKey(kvs[i].Key), kvs[i].Value)
	}

	return nil
}

//查询某个k是否存在
func (l *LevelDB) Exists(key []byte) (int64, error) {
	if err := checkKeySize(key); err != nil {
		return 0, err
	}

	var err error
	key = l.encodeKVKey(key)

	var v []byte
	//snap
	v, err = l.snap.Get(key)

	if err != nil && err.Error() == ErrNotFound.Error() {
		return 0, nil
	}

	if v != nil && err == nil {
		return 1, nil
	}

	return 0, err
}

//批量删除多个k的值
func (l *LevelDB) Del(keys ...[]byte) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	codedKeys := make([][]byte, len(keys))
	for i, k := range keys {
		codedKeys[i] = l.encodeKVKey(k)
	}

	//snap
	l.snap.Del(codedKeys...)
	return 0, nil
}
