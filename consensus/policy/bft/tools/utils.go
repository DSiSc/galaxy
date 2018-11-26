package tools

import "github.com/DSiSc/validator/tools/account"

func AccountFilter(blacklist []account.Account, accounts []account.Account) []account.Account {
	var peer []account.Account
	for _, black := range blacklist {
		peer = filterAccount(black, accounts)
	}
	return peer
}

func filterAccount(black account.Account, accounts []account.Account) []account.Account {
	all := make([]account.Account, 0)
	for _, account := range accounts {
		if black != account {
			all = append(all, account)
		}
	}
	return all
}
