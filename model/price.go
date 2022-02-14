package model

import (
	"gorm.io/gorm"
)

type Price struct {
	gorm.Model
	Fsyms string
	Tsyms string
	Data  string
}

type PriceRepository struct {
	DB *gorm.DB
}

func (pr *PriceRepository) FindByTsymsAndFsyms(fsyms string, tsyms string) *Price {
	var price Price
	res := pr.DB.First(&price, "fsyms = ? and tsyms = ?", fsyms, tsyms)
	if res.Error != nil {
		return nil
	}
	return &price
}

func (pr *PriceRepository) UpsertForTsymsAndFsyms(data string, fsyms string, tsyms string) {
	var price Price
	res := pr.DB.First(&price, "fsyms = ? and tsyms = ?", fsyms, tsyms)

	if res.Error != nil {
		pr.DB.Create(&Price{Data: data, Fsyms: fsyms, Tsyms: tsyms})
	} else {
		pr.DB.Model(&price).Update("data", data)
	}
}
