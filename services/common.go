package services

import "go-service-template/domain"

var (
	SupplierMap = map[int]string{
		1: "Supplier 1",
		2: "Supplier 2",
		3: "Supplier 3",
		4: "Supplier 4",
		5: "Supplier 5",
		6: "Supplier 6",
		7: "Supplier 7",
		8: "Supplier 8",
	}
	LocationTypeMap = map[int]string{
		domain.ReconLocationTypeID:       "Recon",
		domain.WholesaleLocationTypeID:   "Wholesale",
		domain.LastMileLocationTypeID:    "Last Mile",
		domain.RetailReadyLocationTypeID: "Retail Ready",
		domain.CrossDockLocationTypeID:   "Cross Dock",
		domain.StorageLocationTypeID:     "Storage",
		domain.NotOnSiteLocationTypeID:   "Not on Site",
	}
)
