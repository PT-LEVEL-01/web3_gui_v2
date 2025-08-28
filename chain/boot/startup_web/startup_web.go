package startup_web

import (
	"crypto/aes"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"web3_gui/chain/config"
	"web3_gui/keystore/adapter"
	"web3_gui/keystore/adapter/crypto"
	"web3_gui/keystore/adapter/crypto/dh"
	rpc "web3_gui/libp2parea/adapter/sdk/jsonrpc2"
	"web3_gui/libp2parea/adapter/sdk/jsonrpc2/model"
	"web3_gui/libp2parea/adapter/sdk/web"
	"web3_gui/libp2parea/adapter/sdk/web/routers"
	"web3_gui/utils"
)

type startupController struct {
	beego.Controller
	startupSignal chan bool          //启动信号
	firstStartup  bool               //是否首次启动
	started       atomic.Bool        //已经启动标志
	keyStore      *keystore.Keystore //本地keystore
	dhKey         *dh.KeyPair        //服务器密钥对
}

func (i *startupController) Index() {
	handler := &startupHandler{}
	handler.ServeHTTP(i.Ctx.ResponseWriter, i.Ctx.Request)
	return
}

func Start(startupPwd string) (err error) {
	c := &startupController{
		startupSignal: make(chan bool, 1),
		started:       atomic.Bool{},
	}
	c.dhKey, err = updateOrLoadStartupKey(startupPwd)
	if err != nil {
		return err
	}

	c.keyStore = keystore.NewKeystore(config.KeystoreFileAbsPath, config.AddrPre)
	if err = c.keyStore.Load(); err != nil {
		c.firstStartup = true
		err = nil
	}
	web.Start(config.SetLibp2pareaConfig())
	routers.Router("/startup", c, "post:Index")                       //startup
	rpc.RegisterRPC("info", c.info)                                   //获取startup信息
	rpc.RegisterRPC("startup", c.startup)                             //启动
	rpc.RegisterRPC("updateStartupPassword", c.updateStartupPassword) //更新启动密码
	go beego.Run()
	select {
	case <-c.startupSignal: //阻塞至密码验证成功
		return nil
	}
	return nil
}

func (c *startupController) info(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	res, err = model.Tojson(map[string]interface{}{
		"isFirstStartup": c.firstStartup,
		"started":        c.started.Load(),
	})
	return
}

func (c *startupController) startup(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if c.started.Load() {
		res, err = model.Errcode(model.Nomarl, "started")
		return
	}
	clientPukBase64, e := getParamString(rj, "puk")
	if e != nil {
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}
	clientPukBytes, e := base64.StdEncoding.DecodeString(clientPukBase64)
	if e != nil {
		res, err = model.Errcode(model.Nomarl, "invalid base64")
		return
	}
	clientPuk := [32]byte{}
	copy(clientPuk[:], clientPukBytes)
	dbp := dh.NewDHPair(c.dhKey.PrivateKey, clientPuk)
	key := dh.KeyExchange(dbp)
	walletPasswordCipherBase64, e := getParamString(rj, "walletPassword")
	if e != nil {
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}
	walletPassword, e := DecryptCBCBase64(walletPasswordCipherBase64, key)
	if e != nil {
		res, err = model.Errcode(model.Nomarl, "failed to decrypt wallet password")
		return
	}
	fmt.Println("Got WalletPassword:", walletPassword)

	if c.firstStartup {
		repeatWalletPasswordCipherBase64, e := getParamString(rj, "repeatWalletPassword")
		if e != nil {
			res, err = model.Errcode(model.Nomarl, e.Error())
			return
		}
		repeatWalletPassword, e := DecryptCBCBase64(repeatWalletPasswordCipherBase64, key)
		if e != nil {
			res, err = model.Errcode(model.Nomarl, "failed to decrypt repeat wallet password")
			return
		}
		if walletPassword != repeatWalletPassword {
			res, err = model.Errcode(model.Nomarl, "repeat wallet password")
			return
		}
	} else {
		//pwd := sha256.Sum256([]byte(walletPassword)) //钱包的密码
		//验证钱包密码是否正确
		if _, _, e = c.keyStore.GetNetAddr(walletPassword); e != nil {
			res, err = model.Errcode(model.Nomarl, "invalid wallet password")
			return
		}
	}
	config.Wallet_keystore_default_pwd = walletPassword
	c.startupSignal <- true
	c.started.Store(true)
	res, err = model.Tojson("success")
	return
}

func (c *startupController) updateStartupPassword(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	clientPukBase64, e := getParamString(rj, "puk")
	if e != nil {
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}
	clientPukBytes, e := base64.StdEncoding.DecodeString(clientPukBase64)
	if e != nil {
		res, err = model.Errcode(model.Nomarl, "invalid base64")
		return
	}
	clientPuk := [32]byte{}
	copy(clientPuk[:], clientPukBytes)
	dbp := dh.NewDHPair(c.dhKey.PrivateKey, clientPuk)
	key := dh.KeyExchange(dbp)
	startupPasswordCipherBase64, e := getParamString(rj, "startupPassword")
	if e != nil {
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}
	startupPassword, e := DecryptCBCBase64(startupPasswordCipherBase64, key)
	if e != nil {
		res, err = model.Errcode(model.Nomarl, "failed to decrypt startup password")
		return
	}
	if startupPassword == "" {
		res, err = model.Errcode(model.Nomarl, "empty startup password")
		return
	}
	if dhKey, e := updateOrLoadStartupKey(startupPassword); e != nil {
		res, err = model.Errcode(model.Nomarl, "failed to update startup password")
		return
	} else {
		c.dhKey = dhKey
	}
	res, err = model.Tojson("success")
	return
}

func updateOrLoadStartupKey(startupPassword string) (*dh.KeyPair, error) {
	startupfile := filepath.Join(config.Path_configDir, config.Core_startup_key)
	if startupPassword == "" { //load
		b, err := os.ReadFile(startupfile)
		if err != nil {
			return nil, err
		}
		var dhKey = &dh.KeyPair{}
		var block *pem.Block
		for {
			block, b = pem.Decode(b)
			if block == nil {
				break
			}
			if block.Type == "PRIVATE KEY" {
				key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
				if err != nil {
					return nil, err
				}
				prik := key.(ed25519.PrivateKey)
				buf := [32]byte{}
				copy(buf[:], prik.Seed())
				dhKey.PrivateKey = buf
				continue
			}
			if block.Type == "PUBLIC KEY" {
				key, err := x509.ParsePKIXPublicKey(block.Bytes)
				if err != nil {
					return nil, err
				}
				puk := key.(ed25519.PublicKey)
				buf := [32]byte{}
				copy(buf[:], puk)
				dhKey.PublicKey = buf
				continue
			}
		}
		return dhKey, nil
	} else { //createorupdate
		h := sha3.New256()
		h.Write([]byte(startupPassword))
		keyA, err := dh.GenerateKeyPair(h.Sum(nil))
		if err != nil {
			return nil, err
		}
		prikBs, _ := x509.MarshalPKCS8PrivateKey(ed25519.PrivateKey(keyA.PrivateKey[:]))
		prikPem := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: prikBs,
		}
		buf := make([]byte, 0)
		prikPemBytes := pem.EncodeToMemory(prikPem)
		buf = append(buf, prikPemBytes...)
		pukBs, err := x509.MarshalPKIXPublicKey(ed25519.PublicKey(keyA.PublicKey[:]))
		if err != nil {
			return nil, err
		}
		pukPem := &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pukBs,
		}
		pukPemBytes := pem.EncodeToMemory(pukPem)
		buf = append(buf, pukPemBytes...)
		if err := utils.SaveFile(startupfile, &buf); err != nil {
			return nil, err
		}
		return &keyA, nil
	}
}

func EncryptCBCBase64(text string, key [32]byte) (string, error) {
	iv, err := crypto.Rand16Byte()
	if err != nil {
		return "", err
	}
	chiper, err := crypto.EncryptCBC([]byte(text), key[:], iv[:])
	if err != nil {
		return "", err
	}
	chiper = append(iv[:], chiper...)
	return base64.StdEncoding.EncodeToString(chiper), nil
}

func DecryptCBCBase64(chiperBase64 string, key [32]byte) (string, error) {
	chiper, err := base64.StdEncoding.DecodeString(chiperBase64)
	if err != nil {
		return "", err
	}
	text, err := crypto.DecryptCBC(chiper[aes.BlockSize:], key[:], chiper[:aes.BlockSize])
	if err != nil {
		return "", err
	}
	return string(text), nil
}

func getParamString(rj *model.RpcJson, key string) (string, error) {
	param, ok := rj.Get(key)
	if !ok {
		return "", errors.Errorf("missing params %s", key)
	}
	v, ok := param.(string)
	if !ok {
		return "", errors.Errorf("type params %s", key)
	}
	if v == "" {
		return "", errors.Errorf("empty params %s", key)
	}

	return v, nil
}
