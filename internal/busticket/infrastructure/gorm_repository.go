package infrastructure

import (
	"context"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	"github.com/mateusmacedo/go-bff/pkg/application"
)

type gormBusTicketRepository struct {
	db     *gorm.DB
	logger application.AppLogger
}

func NewGormBusTicketRepository(dsn string, logger application.AppLogger) (domain.BusTicketRepository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(&domain.BusTicket{}); err != nil {
		return nil, err
	}

	return &gormBusTicketRepository{
		db:     db,
		logger: logger,
	}, nil
}

func (r *gormBusTicketRepository) Save(ctx context.Context, busTicket domain.BusTicket) error {
	if err := r.db.WithContext(ctx).Create(&busTicket).Error; err != nil {
		application.LogError(ctx, r.logger, "failed to save busTicket", err, map[string]interface{}{
			"busTicket": busTicket,
		})
		return err
	}

	application.LogInfo(ctx, r.logger, "busTicket saved", map[string]interface{}{
		"busTicket": busTicket,
	})
	return nil
}

func (r *gormBusTicketRepository) FindByPassengerName(ctx context.Context, passengerName string) ([]domain.BusTicket, error) {
	var busTickets []domain.BusTicket

	if err := r.db.WithContext(ctx).Where("passenger_name = ?", passengerName).Find(&busTickets).Error; err != nil {
		application.LogError(ctx, r.logger, "failed to find busTickets", err, map[string]interface{}{
			"passengerName": passengerName,
		})
		return nil, err
	}

	application.LogInfo(ctx, r.logger, "busTickets found", map[string]interface{}{
		"passengerName": passengerName,
		"busTickets":    busTickets,
	})

	return busTickets, nil
}

func (r *gormBusTicketRepository) Update(ctx context.Context, busTicket domain.BusTicket) error {
	if err := r.db.WithContext(ctx).Model(&domain.BusTicket{}).Where("id = ?", busTicket.ID).Updates(busTicket).Error; err != nil {
		application.LogError(ctx, r.logger, "failed to update busTicket", err, map[string]interface{}{
			"busTicket": busTicket,
		})
		return err
	}

	application.LogInfo(ctx, r.logger, "busTicket updated", map[string]interface{}{
		"busTicket": busTicket,
	})

	return nil
}
