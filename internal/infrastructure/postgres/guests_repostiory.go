package postgres

import (
	"context"
	"errors"
	domain "metrika/internal/domain/analytics"

	"gorm.io/gorm"
)

type GuestsRepository struct {
	db *gorm.DB
}

func NewGuestsRepository(db *gorm.DB) *GuestsRepository {
	return &GuestsRepository{db}
}

func (r *GuestsRepository) ByFingerprint(ctx context.Context, fingerprint string) (*domain.Guest, error) {
	db := getDB(ctx, r.db)

	var mGuest Guest

	if err := db.Model(&Guest{}).Where("f_id = ?", fingerprint).First(&mGuest).Error; err != nil {
		return nil, err
	}

	return &domain.Guest{Fingerprint: fingerprint, ID: mGuest.ID, DomainID: mGuest.DomainID}, nil
}

func (r *GuestsRepository) FirstOrCreate(ctx context.Context, fingerprint string, domain_id uint) (*domain.Guest, error) {
	db := getDB(ctx, r.db)

	mGuest := Guest{
		Fingerprint: fingerprint,
		DomainID:    domain_id,
	}

	if err := db.Model(&Guest{}).Where("f_id = ?", fingerprint).FirstOrCreate(&mGuest).Error; err != nil {
		return nil, err
	}

	return &domain.Guest{DomainID: domain_id, Fingerprint: fingerprint, ID: mGuest.ID}, nil
}

func (r *GuestsRepository) CreateGuests(ctx context.Context, guests *[]domain.Guest) ([]domain.Guest, error) {
	db := getDB(ctx, r.db)

	var mGuests []Guest
	for _, guest := range *guests {
		mGuests = append(mGuests, Guest{
			Fingerprint: guest.Fingerprint,
			DomainID:    guest.DomainID,
		})
	}

	if err := db.Model(&Guest{}).Create(&mGuests).Error; err != nil {
		return nil, err
	}

	var dGuests []domain.Guest
	for _, guest := range mGuests{
		dGuests = append(dGuests, domain.Guest{
			ID: guest.ID,
			DomainID: guest.DomainID,
			Fingerprint: guest.Fingerprint,
		})
	}

	return dGuests, nil
}


//TODO:доделать, тут надо искать start date и end date у сессий, а domain_id - у посетителя
func(r *GuestsRepository) Find(ctx context.Context, opts domain.FindGuestsOptions) (*[]domain.Guest, error){
	db := getDB(ctx, r.db)

	var mGuests *[]Guest

	query := db.Model(&Guest{}).Joins("LEFT JOIN guest_sessions gs ON guests.id=gs.guest_id").Where("guests.domain_id = ?", opts.DomainID)

	if opts.StartDate != nil{
		query.Where("NOT gs.created_at < ?", opts.StartDate)
	}
	if opts.EndDate != nil{
		query.Where("NOT gs.created_at > ?", opts.EndDate)
	}
	if opts.Limit != nil{
		query.Limit(*opts.Limit)
	}
	if opts.Offset != nil{
		query.Offset(*opts.Offset)
	}

	if err := query.Find(&mGuests).Error; err != nil{
		if errors.Is(err, gorm.ErrRecordNotFound){
			return nil, domain.ErrGuestsNotFound
		}
		return nil, err
	}

	var guests []domain.Guest

	for _, guest := range *mGuests{
		//todo:запрос говно. надо переделать. 
		//todo:посмотреть как вообще парсятся данные с результата sql запросов, про подзапросы, как можно форматировать вид результатов запроса
		//todo: например чтобы в результате запроса само выводилось firstVisit и LastVisit, а не пришлось бы его высчитывать в коде
		guests = append(guests, domain.Guest{
			ID: guest.ID,
			DomainID: guest.DomainID,
			// FirstVisit: ,
		})
	}

}