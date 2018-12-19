package utils

import (
	"github.com/DSiSc/validator/tools/account"
)

func Abs(x uint64, y uint64) int64 {
	n := int64(x - y)
	z := n >> 63
	return (n ^ z) - z
}

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

func GetAccountById(peers []account.Account, expect uint64) account.Account {
	var temp account.Account
	for _, peer := range peers {
		if peer.Extension.Id == expect {
			temp = peer
		}
	}
	return temp
}

func GetAccountWithMinId(accounts []account.Account) account.Account {
	var minNode = accounts[0]
	for _, node := range accounts {
		if node.Extension.Id < minNode.Extension.Id {
			minNode = node
		}
	}
	return minNode
}
