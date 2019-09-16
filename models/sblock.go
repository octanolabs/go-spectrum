package models

type Sblock struct {
	Number       uint64 `bson:"number" json:"number"`
	Hash         string `bson:"hash" json:"hash"`
	Timestamp    uint64 `bson:"timestamp" json:"timestamp"`
	BlockReward  string `bson:"blockReward" json:"blockReward"`
	UncleRewards string `bson:"uncleRewards" json:"uncleRewards"`
	Minted       string `bson:"minted" json:"minted"`
	Supply       string `bson:"supply" json:"supply"`
}
