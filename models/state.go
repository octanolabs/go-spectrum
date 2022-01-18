package models

type RawState struct {
	Root     string                 `bson:"root" json:"root"`
	Accounts map[string]interface{} `bson:"accounts" json:"accounts"`
}

type Account struct {
	Address string `bson:"address" json:"address"`
	Balance string `bson:"balance" json:"balance"`
	Block   uint64 `bson:"block" json:"block"`
}
