# pkg

```go
package infrastructure

import (
 "context"
 "errors"
 "sync"

 "github.com/mateusmacedo/go-bff/pkg/application"
 "github.com/mateusmacedo/go-bff/pkg/domain"
)

// simpleQueryBus é uma implementação simples de um barramento de consultas.
type simpleQueryBus[Q domain.Query[T], T any, R any] struct {
 handlers map[string]application.QueryHandler[Q, T, R]
 mu       sync.RWMutex
}

// NewSimpleQueryBus cria uma nova instância de SimpleQueryBus.
func NewSimpleQueryBus[Q domain.Query[T], T any, R any]() *simpleQueryBus[Q, T, R] {
 return &simpleQueryBus[Q, T, R]{
  handlers: make(map[string]application.QueryHandler[Q, T, R]),
 }
}

// RegisterHandler registra um manipulador para uma consulta específica.
func (bus *simpleQueryBus[Q, T, R]) RegisterHandler(queryName string, handler application.QueryHandler[Q, T, R]) {
 bus.mu.Lock()
 defer bus.mu.Unlock()
 bus.handlers[queryName] = handler
}

// Dispatch despacha uma consulta para o manipulador registrado usando goroutines.
func (bus *simpleQueryBus[Q, T, R]) Dispatch(ctx context.Context, query Q) (R, error) {
 bus.mu.RLock()
 handler, found := bus.handlers[query.QueryName()]
 bus.mu.RUnlock()

 var zero R
 if !found {
  return zero, errors.New("no handler registered for query")
 }

 resultChan := make(chan R, 1)
 errChan := make(chan error, 1)

 go func() {
  result, err := handler.Handle(ctx, query)
  if err != nil {
   errChan <- err
   return
  }
  resultChan <- result
 }()

 select {
 case <-ctx.Done():
  return zero, ctx.Err()
 case result := <-resultChan:
  return result, nil
 case err := <-errChan:
  return zero, err
 }
}
```

## Internal

```go
// internal/domain/passage.go
package domain

import "time"

// Passage representa uma passagem rodoviária.
type Passage struct {
 ID            string
 PassengerName string
 DepartureTime time.Time
 SeatNumber    int
 Origin        string
 Destination   string
}

// PassageRepository define a interface para o repositório de passagens.
type PassageRepository interface {
 Save(passage Passage) error
 FindByID(id string) (Passage, error)
 Update(passage Passage) error
}

// internal/application/command.go
package application

import (
 "time"

 "github.com/mateusmacedo/go-bff/pkg/domain"
)

// ReservePassageData contém os dados necessários para reservar uma passagem.
type ReservePassageData struct {
 PassengerName string
 DepartureTime time.Time
 SeatNumber    int
 Origin        string
 Destination   string
}

// reservePassageCommand é uma implementação privada de um comando para reservar uma passagem.
type reservePassageCommand struct {
 data ReservePassageData
}

func (c reservePassageCommand) CommandName() string {
 return "ReservePassage"
}

func (c reservePassageCommand) Payload() ReservePassageData {
 return c.data
}

// NewReservePassageCommand cria um novo comando para reservar uma passagem.
func NewReservePassageCommand(data ReservePassageData) domain.Command[ReservePassageData] {
 return reservePassageCommand{data: data}
}

// internal/application/event.go
package application

import (
 "github.com/mateusmacedo/go-bff/pkg/domain"
)

// PassageBookedEvent representa um evento de passagem reservada.
type passageBookedEvent struct {
 data string
}

func (e passageBookedEvent) EventName() string {
 return "PassageBooked"
}

func (e passageBookedEvent) Payload() string {
 return e.data
}

// NewPassageBookedEvent cria um novo evento de passagem reservada.
func NewPassageBookedEvent(data string) domain.Event[string] {
 return passageBookedEvent{data: data}
}

// internal/application/handler.go
package application

import (
 "context"

 "github.com/mateusmacedo/go-bff/internal/domain"
 pkgApp "github.com/mateusmacedo/go-bff/pkg/application"
 pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
)

// reservePassageHandler manipula o comando de reserva de passagem.
type reservePassageHandler struct {
 repository  domain.PassageRepository
 idGenerator pkgDomain.IDGenerator[string]
}

func (h *reservePassageHandler) Handle(ctx context.Context, command pkgDomain.Command[ReservePassageData]) error {
 select {
 case <-ctx.Done():
  return ctx.Err()
 default:
  data := command.Payload()
  passage := domain.Passage{
   ID:            h.idGenerator(),
   PassengerName: data.PassengerName,
   DepartureTime: data.DepartureTime,
   SeatNumber:    data.SeatNumber,
   Origin:        data.Origin,
   Destination:   data.Destination,
  }

  if err := h.repository.Save(passage); err != nil {
   return err
  }

  // Aqui você pode publicar um evento de passagem reservada.
  return nil
 }
}

// NewReservePassageHandler cria um novo handler para o comando de reserva de passagem.
func NewReservePassageHandler(repo domain.PassageRepository) pkgApp.CommandHandler[pkgDomain.Command[ReservePassageData], ReservePassageData] {
 return &reservePassageHandler{
  repository: repo,
 }
}

// findPassageHandler manipula a consulta para encontrar uma passagem.
type findPassageHandler struct {
 repository domain.PassageRepository
}

func (h *findPassageHandler) Handle(ctx context.Context, query pkgDomain.Query[FindPassageData]) (domain.Passage, error) {
 select {
 case <-ctx.Done():
  return domain.Passage{}, ctx.Err()
 default:
  data := query.Payload()
  return h.repository.FindByID(data.PassageID)
 }
}

// NewFindPassageHandler cria um novo handler para a consulta de encontrar passagem.
func NewFindPassageHandler(repo domain.PassageRepository) pkgApp.QueryHandler[pkgDomain.Query[FindPassageData], FindPassageData, domain.Passage] {
 return &findPassageHandler{
  repository: repo,
 }
}

// internal/application/query.go
package application

import (
 "github.com/mateusmacedo/go-bff/pkg/domain"
)

// FindPassageData contém os dados necessários para encontrar uma passagem.
type FindPassageData struct {
 PassageID string
}

// findPassageQuery é uma implementação privada de uma consulta para encontrar uma passagem.
type findPassageQuery struct {
 data FindPassageData
}

func (q findPassageQuery) QueryName() string {
 return "FindPassage"
}

func (q findPassageQuery) Payload() FindPassageData {
 return q.data
}

// NewFindPassageQuery cria uma nova consulta para encontrar uma passagem.
func NewFindPassageQuery(data FindPassageData) domain.Query[FindPassageData] {
 return findPassageQuery{data: data}
}

// internal/infrastructure/repository.go
package infrastructure

import (
 "errors"

 "github.com/mateusmacedo/go-bff/internal/domain"
)

// InMemoryPassageRepository é uma implementação em memória do repositório de passagens.
type InMemoryPassageRepository struct {
 data map[string]domain.Passage
}

func NewInMemoryPassageRepository() *InMemoryPassageRepository {
 return &InMemoryPassageRepository{
  data: make(map[string]domain.Passage),
 }
}

func (r *InMemoryPassageRepository) Save(passage domain.Passage) error {
 if _, exists := r.data[passage.ID]; exists {
  return errors.New("passage already exists")
 }
 r.data[passage.ID] = passage
 return nil
}

func (r *InMemoryPassageRepository) FindByID(id string) (domain.Passage, error) {
 passage, exists := r.data[id]
 if !exists {
  return domain.Passage{}, errors.New("passage not found")
 }
 return passage, nil
}

func (r *InMemoryPassageRepository) Update(passage domain.Passage) error {
 if _, exists := r.data[passage.ID]; !exists {
  return errors.New("passage not found")
 }
 r.data[passage.ID] = passage
 return nil
}
```

### cmd

```go
// cmd/main.go
package main

import (
 "context"
 "fmt"
 "time"

 "github.com/google/uuid"

 "github.com/mateusmacedo/go-bff/internal/application"
 "github.com/mateusmacedo/go-bff/internal/domain"
 "github.com/mateusmacedo/go-bff/internal/infrastructure"
 pkgDomain "github.com/mateusmacedo/go-bff/pkg/domain"
 pkgInfra "github.com/mateusmacedo/go-bff/pkg/infrastructure"
)

func main() {
 // Configuração do repositório
 repository := infrastructure.NewInMemoryPassageRepository()

 // Gerador de ID
 idGenerator := func() string {
  return uuid.New().String()
 }

 // Criação dos handlers
 reserveHandler := application.NewReservePassageHandler(repository, idGenerator)
 findHandler := application.NewFindPassageHandler(repository)

 // Criação dos barramentos
 commandBus := pkgInfra.NewSimpleCommandBus[pkgDomain.Command[application.ReservePassageData], application.ReservePassageData]()
 queryBus := pkgInfra.NewSimpleQueryBus[pkgDomain.Query[application.FindPassageData], application.FindPassageData, domain.Passage]()
 eventBus := pkgInfra.NewSimpleEventBus[pkgDomain.Event[string], string]()

 // Registro dos handlers nos barramentos
 commandBus.RegisterHandler("ReservePassage", reserveHandler)
 queryBus.RegisterHandler("FindPassage", findHandler)

 // Criando um contexto com timeout
 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
 defer cancel()

 // Criando um comando de reserva de passagem
 reserveData := application.ReservePassageData{
  PassengerName: "John Doe",
  DepartureTime: time.Now().Add(24 * time.Hour),
  SeatNumber:    12,
  Origin:        "City A",
  Destination:   "City B",
 }
 command := application.NewReservePassageCommand(reserveData)

 // Despachando o comando
 if err := commandBus.Dispatch(ctx, command); err != nil {
  fmt.Println("Erro ao reservar passagem:", err)
  return
 }
 fmt.Println("Passagem reservada com sucesso!")

 // Obtendo o ID da passagem diretamente do repositório para evitar inconsistências
 var passageID string
 for id, passage := range repository.GetData() { // Utilize GetData() se disponível
  if passage.PassengerName == reserveData.PassengerName && passage.DepartureTime.Equal(reserveData.DepartureTime) {
   passageID = id
   break
  }
 }

 // Verifique se o ID foi encontrado
 if passageID == "" {
  fmt.Println("Erro: ID da passagem não encontrado após a reserva.")
  return
 }

 // Criando uma consulta para encontrar uma passagem
 query := application.NewFindPassageQuery(application.FindPassageData{
  PassageID: passageID, // Use o ID gerado corretamente
 })

 // Despachando a consulta
 passage, err := queryBus.Dispatch(ctx, query)
 if err != nil {
  fmt.Println("Erro ao encontrar passagem:", err)
 } else {
  fmt.Printf("Passagem encontrada: %+v\n", passage)
 }

 // Exemplo de publicação de um evento
 event := application.NewPassageBookedEvent("Passage successfully booked for John Doe")
 if err := eventBus.Publish(ctx, event); err != nil {
  fmt.Println("Erro ao publicar evento:", err)
 }
}
```
