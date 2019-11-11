package storage

//
//import (
//	"time"
//
//	"github.com/globalsign/mgo"
//	"github.com/globalsign/mgo/bson"
//	"github.com/ubiq/spectrum-backend/models"
//)
//
///* Chart iterators */
//
//var EOD = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 23, 59, 59, 0, time.UTC)
//
//func (m *MongoDB) GetTxnCounts(days int) *mgo.Iter {
//	var from int64
//
//	hours, _ := time.ParseDuration("-23h59m59s")
//
//	if days == 0 {
//		from = 1485633600
//	} else {
//		from = EOD.Add(hours).AddDate(0, 0, -days).Unix()
//	}
//
//	pipeline := []bson.M{{"$match": bson.M{"timestamp": bson.M{"$gte": from}}}}
//
//	pipe := m.db.C(models.TXNS).Pipe(pipeline)
//
//	return pipe.Iter()
//
//}
//
//func (m *MongoDB) GetBlocks(days int) *mgo.Iter {
//	// genesis block: 1485633600
//	var from int64
//
//	hours, _ := time.ParseDuration("-23h59m59s")
//
//	if days == 0 {
//		from = 1485633600
//	} else {
//		from = EOD.Add(hours).AddDate(0, 0, -days).Unix()
//	}
//
//	pipeline := []bson.M{{"$match": bson.M{"timestamp": bson.M{"$gte": from}}}, {"$sort": bson.M{"number": -1}}}
//
//	pipe := m.db.C(models.BLOCKS).Pipe(pipeline)
//
//	return pipe.Iter()
//
//}
//
//func (m *MongoDB) GetTokenTransfers(contractAddress, address string, after int64) *mgo.Iter {
//
//	pipeline := []bson.M{}
//
//	if contractAddress != "" && address != "" {
//		pipeline = []bson.M{{"$match": bson.M{"contract": contractAddress}}, {"$match": bson.M{"timestamp": bson.M{"$gte": after}}}, {"$match": bson.M{"from": address}}}
//	} else {
//		pipeline = []bson.M{{"$sort": bson.M{"timestamp": 1}}}
//	}
//
//	pipe := m.db.C(models.TRANSFERS).Pipe(pipeline)
//
//	return pipe.Iter()
//
//}
//
//func (m *MongoDB) BlocksIter(blockno uint64) *mgo.Iter {
//
//	pipeline := []bson.M{{"$match": bson.M{"number": bson.M{"$gte": blockno}}}, {"$sort": bson.M{"number": 1}}}
//
//	pipe := m.db.C(models.BLOCKS).Pipe(pipeline)
//
//	return pipe.Iter()
//
//}
