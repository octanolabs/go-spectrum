package storage

import (
	"github.com/octanolabs/go-spectrum/models"
)

func (m *MongoDB) AddSupplyBlock(b models.Sblock) error {
	ss := m.db.C(models.SBLOCK)

	if err := ss.Insert(b); err != nil {
		return err
	}
	return nil
}
