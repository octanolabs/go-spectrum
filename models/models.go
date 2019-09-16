package models

const (
	BLOCKS = "blocks"
	STORE  = "sysstores"
	SBLOCK = "sblock"
)

type Store struct {
	Timestamp int64     `bson:"timestamp" json:"timestamp"`
	Symbol    string    `bson:"symbol" json:"symbol"`
	Sync      [1]uint64 `bson:"sync"`
}
