package TronGrid

import (
	"fmt"
	"testing"
)

func TestGetAccountBalance(t *testing.T) {
	balance, err := GetAccountBalance("TUoHaVjx7n5xz8LwPRDckgFrDWhMhuSuJM")
	if err != nil {
		panic(err)
	}
	fmt.Println("地址余额", balance)
}

func TestCreateAccount(t *testing.T) {
	account, err := CreateAccount()
	if err != nil {
		panic(err)
	}
	fmt.Println("新地址", account)
}
