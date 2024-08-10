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
		return logError(ctx, r.logger, "failed to save busTicket", err, busTicket)
	}

	r.logger.Info(ctx, "busTicket saved", map[string]interface{}{
		"busTicket": busTicket,
	})
	return nil
}

func (r *gormBusTicketRepository) FindByPassengerName(ctx context.Context, passengerName string) ([]domain.BusTicket, error) {
	var busTickets []domain.BusTicket

	if err := r.db.WithContext(ctx).Where("passenger_name = ?", passengerName).Find(&busTickets).Error; err != nil {
		return nil, logError(ctx, r.logger, "failed to find busTickets", err, passengerName)
	}

	r.logger.Info(ctx, "busTickets found", map[string]interface{}{
		"busTickets": busTickets,
	})
	return busTickets, nil
}

func (r *gormBusTicketRepository) Update(ctx context.Context, busTicket domain.BusTicket) error {
	if err := r.db.WithContext(ctx).Model(&domain.BusTicket{}).Where("id = ?", busTicket.ID).Updates(busTicket).Error; err != nil {
		return logError(ctx, r.logger, "failed to update busTicket", err, busTicket)
	}

	r.logger.Info(ctx, "busTicket updated", map[string]interface{}{
		"busTicket": busTicket,
	})
	return nil
}

func logError(ctx context.Context, logger pkgApp.AppLogger, message string, err error, details interface{}) error {
	logger.Error(ctx, message, map[string]interface{}{
		"details": details,
		"error":   err,
	})
	return err
}
