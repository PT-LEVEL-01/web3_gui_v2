package startup_web

import (
	"bytes"
	"encoding/base64"
	"testing"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/adapter/crypto/dh"
)

func Test_Startup(t *testing.T) {
	dhKey, err := updateOrLoadStartupKey("111")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("server prik:", dhKey.PrivateKey)
	t.Log("server puk:", dhKey.PublicKey)
	seed, _ := crypto.Rand32Byte()
	clientKey, _ := dh.GenerateKeyPair(seed[:])
	t.Log("client prik:", clientKey.PrivateKey)
	t.Log("client puk:", clientKey.PublicKey)
	t.Log("client puk base64:", base64.StdEncoding.EncodeToString(clientKey.PublicKey[:]))
	dbpA := dh.NewDHPair(dhKey.PrivateKey, clientKey.PublicKey)
	keyA := dh.KeyExchange(dbpA)
	dbpB := dh.NewDHPair(clientKey.PrivateKey, dhKey.PublicKey)
	keyB := dh.KeyExchange(dbpB)
	t.Log("key:", keyA, bytes.Equal(keyA[:], keyB[:]))
	password := "123456789"
	chiper, _ := EncryptCBCBase64(password, keyA)
	text, _ := DecryptCBCBase64(chiper, keyB)
	t.Log("chiper:", text, chiper)
	password = "111"
	chiper, _ = EncryptCBCBase64(password, keyA)
	text, _ = DecryptCBCBase64(chiper, keyB)
	t.Log("chiper:", text, chiper)
	password = "222"
	chiper, _ = EncryptCBCBase64(password, keyA)
	text, _ = DecryptCBCBase64(chiper, keyB)
	t.Log("chiper:", text, chiper)
}
