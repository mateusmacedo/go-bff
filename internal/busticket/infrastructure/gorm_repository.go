package infrastructure

import (
	"context"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	pkgApp "github.com/mateusmacedo/go-bff/pkg/application"
)

type gormBusTicketRepository struct {
	db     *gorm.DB
	logger pkgApp.AppLogger
}

func NewGormBusTicketRepository(dsn string, logger pkgApp.AppLogger) (domain.BusTicketRepository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate the BusTicket schema
	err = db.AutoMigrate(&domain.BusTicket{})
	if err != nil {
		return nil, err
	}

	return &gormBusTicketRepository{
		db:     db,
		logger: logger,
	}, nil
}

func (r *gormBusTicketRepository) Save(ctx context.Context, busTicket domain.BusTicket) error {
	if err := r.db.WithContext(ctx).Create(&busTicket).Error; err != nil {
		r.logger.Error(ctx, "failed to save busTicket", map[string]interface{}{
			"busTicket": busTicket,
			"error":     err,
		})
		return err
	}

	r.logger.Info(ctx, "busTicket saved", map[string]interface{}{
		"busTicket": busTicket,
	})
	return nil
}

func (r *gormBusTicketRepository) FindByPassengerName(ctx context.Context, passengerName string) ([]domain.BusTicket, error) {
	var busTickets []domain.BusTicket

	if err := r.db.WithContext(ctx).Where("passenger_name = ?", passengerName).Find(&busTickets).Error; err != nil {
		r.logger.Error(ctx, "failed to find busTickets", map[string]interface{}{
			"passengerName": passengerName,
			"error":         err,
		})
		return nil, err
	}

	r.logger.Info(ctx, "busTickets found", map[string]interface{}{
		"busTickets": busTickets,
	})

	return busTickets, nil
}

func (r *gormBusTicketRepository) Update(ctx context.Context, busTicket domain.BusTicket) error {
	if err := r.db.WithContext(ctx).Model(&domain.BusTicket{}).Where("id = ?", busTicket.ID).Updates(busTicket).Error; err != nil {
		r.logger.Error(ctx, "failed to update busTicket", map[string]interface{}{
			"busTicket": busTicket,
			"error":     err,
		})
		return err
	}

	r.logger.Info(ctx, "busTicket updated", map[string]interface{}{
		"busTicket": busTicket,
	})
	return nil
}
