package storage

import "github.com/octanolabs/go-spectrum/models"

func (m *MongoDB) AddTransaction(tx *models.Transaction) error {
	ss := m.db.C(models.TXNS)

	if err := ss.Insert(tx); err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) AddTokenTransfer(tt *models.TokenTransfer) error {
	ss := m.db.C(models.TRANSFERS)

	if err := ss.Insert(tt); err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) AddUncle(u *models.Uncle) error {
	ss := m.db.C(models.UNCLES)

	if err := ss.Insert(u); err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) AddBlock(b *models.Block) error {
	ss := m.db.C(models.BLOCKS)

	if err := ss.Insert(b); err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) AddForkedBlock(b *models.Block) error {
	ss := m.db.C(models.REORGS)

	if err := ss.Insert(b); err != nil {
		return err
	}
	return nil
}
