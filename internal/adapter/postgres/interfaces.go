package postgres

import "roboticCrewChallenge/internal/domain"

var (
	_ domain.PetRepository      = (*PetRepository)(nil)
	_ domain.MerchantRepository = (*MerchantRepository)(nil)
	_ domain.StoreRepository    = (*StoreRepository)(nil)
	_ domain.CustomerRepository = (*CustomerRepository)(nil)
)
