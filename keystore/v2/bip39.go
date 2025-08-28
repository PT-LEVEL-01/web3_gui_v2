package keystore

import "github.com/tyler-smith/go-bip39"

/*
BIP39是用于生成BIP32种子的规范的名称。

BIP39，全称Bitcoin Improvement Proposal 39，中文名为助记词的生成方法，是比特币的一项改进提议。

BIP39常用于生成BIP32的种子。它规定了如何将随机数转换为一组易记的单词，这就是我们经常看到的一组12-24字的备份助记词。当用户生成钱包或首次打开钱包应用程序时，
就会创建这样一组单词。在大部分现代加密货币的钱包中，BIP39都被广泛地使用。

以下是关于BIP39更详细的介绍：

生成方法：助记词是由一组英文单词组成，这些单词都来自固定的单词表（共2048个）中。首先，根据用户的随机动作（如鼠标移动、键盘打字速度等）生成一个随机数，
        然后将这个随机数转换为一组单词。这个过程是可逆的，只要记住这组单词，就能恢复原始的随机数。
密钥恢复：在BIP39规定的助记词生成方法中，只需记住12-24个简单的单词，用户就可以恢复整个钱包。这极大地简化了钱包私钥的备份和恢复问题，提高了钱包的可用性。
密码功能：BIP39规定的密码功能可以增加恢复私钥的难度，使别人更难通过助记词恢复到你的私钥。一旦设置了BIP39密码，就必须通过助记词+密码的方式来恢复钱包。
跨钱包兼容性：由于BIP39是一个开放的标准，各个钱包厂商之间有很高的兼容性，这意味着你可以把一个钱包的助记词导入到另一个钱包中，这极大地方便了用户。
BIP39是比特币钱包中一个非常重要的标准，它通过助记词将复杂的私钥管理问题简化，使得用户更加容易使用比特币。
*/

type MnemonicSize int

const (
	MnemonicSize128 MnemonicSize = 128 //12个单词
	MnemonicSize256 MnemonicSize = 256 //24个单词
)

/*
随机创建一个助记词
*/
func CreateMnemonic(bitSize MnemonicSize) (string, error) {
	entropy, err := bip39.NewEntropy(int(bitSize)) //128 12位 256 24位.
	if err != nil {
		return "", err
	}
	Mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}
	return Mnemonic, nil
}
