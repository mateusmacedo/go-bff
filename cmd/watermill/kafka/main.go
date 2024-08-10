package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mateusmacedo/go-bff/internal/busticket"
	"github.com/mateusmacedo/go-bff/internal/busticket/application"
	"github.com/mateusmacedo/go-bff/internal/busticket/domain"
	"github.com/mateusmacedo/go-bff/internal/busticket/infrastructure"
	pkgApp "github.com/mateusmacedo/go-bff/pkg/application"
	pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
	"github.com/mateusmacedo/go-bff/pkg/infrastructure/kafka/adapter"
	watermillLogAdapter "github.com/mateusmacedo/go-bff/pkg/infrastructure/watermill/adapter"
	zapAdapter "github.com/mateusmacedo/go-bff/pkg/infrastructure/zaplogger/adapter"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	appLogger, err := zapAdapter.NewZapAppLogger()
	if err != nil {
		panic(err)
	}

	logger := watermillLogAdapter.NewWatermillLoggerAdapter(appLogger)
	marshaler := kafka.DefaultMarshaler{}

	publisher, err := createKafkaPublisher(logger, marshaler)
	if err != nil {
		appLogger.Error(ctx, "Erro ao criar publisher", map[string]interface{}{"error": err})
		return
	}
	defer publisher.Close()

	subscriber, err := createKafkaSubscriber(logger, marshaler)
	if err != nil {
		appLogger.Error(ctx, "Erro ao criar subscriber", map[string]interface{}{"error": err})
		return
	}
	defer subscriber.Close()

	commandBus := adapter.NewKafkaCommandBus[pkgDomain.Command[application.ReserveBusTicketData], application.ReserveBusTicketData](publisher, subscriber, appLogger)
	queryBus := adapter.NewKafkaQueryBus[pkgDomain.Query[application.FindBusTicketData], application.FindBusTicketData, []domain.BusTicket](publisher, subscriber, appLogger)
	eventBus := adapter.NewKafkaEventBus[pkgDomain.Event[string], string](publisher, subscriber, appLogger)

	idGenerator := uuid.NewString

	dsn := "host=localhost user=myuser password=mypassword dbname=mydb port=5432 sslmode=disable TimeZone=UTC"
	busTicketRepo, err := infrastructure.NewGormBusTicketRepository(dsn, appLogger)
	if err != nil {
		appLogger.Error(ctx, "Erro ao inicializar o repositório", map[string]interface{}{"error": err})
		panic(err)
	}

	busTicketSlice := busticket.NewBusTicketSlice(commandBus, queryBus, idGenerator, appLogger, eventBus, busTicketRepo)
	router := chi.NewRouter()
	busTicketSlice.RegisterRoutes(router)

	go handleShutdown(ctx, cancel, appLogger)

	serverAddress := ":8080"
	server := &http.Server{Addr: serverAddress, Handler: router}

	go startServer(ctx, server, appLogger, serverAddress)

	<-ctx.Done()
	shutdownServer(ctx, server, appLogger)
}

func createKafkaPublisher(logger watermill.LoggerAdapter, marshaler kafka.DefaultMarshaler) (*kafka.Publisher, error) {
	publisherConfig := kafka.PublisherConfig{
		Brokers:   []string{"localhost:9092"},
		Marshaler: marshaler,
	}
	return kafka.NewPublisher(publisherConfig, logger)
}

func createKafkaSubscriber(logger watermill.LoggerAdapter, marshaler kafka.DefaultMarshaler) (*kafka.Subscriber, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V1_0_0_0
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.ClientID = "watermill"

	subscriberConfig := kafka.SubscriberConfig{
		Brokers:               []string{"localhost:9092"},
		Unmarshaler:           marshaler,
		ConsumerGroup:         "example_consumer_group",
		OverwriteSaramaConfig: saramaConfig,
		InitializeTopicDetails: &sarama.TopicDetail{
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}
	return kafka.NewSubscriber(subscriberConfig, logger)
}

func handleShutdown(ctx context.Context, cancel context.CancelFunc, appLogger pkgApp.AppLogger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	appLogger.Info(ctx, "Sinal capturado", map[string]interface{}{"signal": sig})
	cancel()
}

func startServer(ctx context.Context, server *http.Server, appLogger pkgApp.AppLogger, serverAddress string) {
	appLogger.Info(ctx, "Server starting on:"+serverAddress, nil)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		appLogger.Error(ctx, "Erro ao iniciar o servidor", map[string]interface{}{"error": err})
	}
}

func shutdownServer(ctx context.Context, server *http.Server, appLogger pkgApp.AppLogger) {
	appLogger.Info(ctx, "Encerrando servidor...", nil)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Error(context.Background(), "Erro ao encerrar servidor", map[string]interface{}{"error": err})
	}
	appLogger.Info(context.Background(), "Servidor encerrado", nil)
}
