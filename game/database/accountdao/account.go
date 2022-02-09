package accountdao

import (
	t "game/typedefine"
)

var (
	accounts    = make(map[string]*t.Account)
)

func AddAccount(account *t.Account) {
	accounts[account.AccountId] = account
}

func GetAccount(accountId string) *t.Account {
	return accounts[accountId]
}

func DelAccount(accountId string) {
	delete(accounts, accountId)
}
