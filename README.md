# Descrição do Projeto

O projeto é uma aplicação de reserva de passagens rodoviárias desenvolvida em Go. Utiliza uma arquitetura baseada em mensagens, aproveitando o Watermill para comunicação assíncrona entre componentes. O sistema implementa padrões de design como *CQRS* (Command Query Responsibility Segregation) para separar comandos de consulta e manipulação de eventos, o que facilita a escalabilidade e manutenibilidade.

## Propósito do Projeto

O principal objetivo deste projeto é demonstrar como criar um sistema de reservas utilizando uma arquitetura moderna e escalável, que permita fácil adaptação e extensão para uso em ambientes de produção. Ele serve como um exemplo de como implementar um sistema orientado a mensagens, permitindo a integração com diferentes *brokers* de mensagens (como Kafka, RabbitMQ ou Redis) com mudanças mínimas.

### Componentes Reutilizáveis

#### 1. **Pacote `domain`**

- **Interfaces Comuns**: Define interfaces para `Command`, `Event`, e `Query`, que podem ser reutilizadas em outros contextos de negócios para padronizar como os comandos e consultas são representados.

- **Modelo de Domínio `Passage`**: Representa a entidade `Passage` e pode ser facilmente adaptado para diferentes contextos de domínio.

#### 2. **Pacote `application`**

- **CommandBus, QueryBus, EventBus**: Interfaces para gerenciamento de comandos, consultas e eventos. Permitem o desacoplamento entre a lógica de negócios e a infraestrutura de comunicação.

- **Handlers de Comando e Consulta**: Implementações que manipulam a lógica de negócios para reserva e consulta de passagens.

- **Logger Abstrato**: Define uma interface para logging que pode ser implementada de diferentes maneiras, conforme necessário.

#### 3. **Pacote `infrastructure`**

- **InMemoryPassageRepository**: Um repositório em memória para armazenar passagens, útil para testes e simulações. Pode ser substituído por implementações baseadas em banco de dados.

- **Adaptadores Watermill**: Implementações de `CommandBus`, `QueryBus` e `EventBus` usando o Watermill. Podem ser facilmente adaptadas para usar diferentes *brokers* de mensagens.

- **Gerador de UUID**: Função para gerar UUIDs, útil em muitos contextos de geração de identificadores únicos.

### Contexto Interno

#### Arquitetura

A arquitetura do sistema é baseada no padrão CQRS e utiliza o Watermill para comunicação entre componentes através de mensagens. Isso permite um sistema altamente desacoplado, onde comandos e consultas são processados independentemente, e eventos são publicados para notificar outras partes do sistema sobre alterações de estado.

#### Fluxo de Mensagens

1. **Comandos**: São enviados para o `CommandBus`, que os despacha para o handler correspondente.
   - Exemplo: `ReservePassageCommand` é enviado para reservar uma passagem.

2. **Consultas**: São enviadas para o `QueryBus`, que retorna o resultado após consultar o handler adequado.
   - Exemplo: `FindPassageQuery` é usado para recuperar uma passagem específica.

3. **Eventos**: São publicados no `EventBus` para informar outras partes do sistema sobre ações concluídas.
   - Exemplo: `PassageBookedEvent` notifica que uma passagem foi reservada com sucesso.

#### Camadas do Sistema

- **Domain Layer**: Define as interfaces e estruturas de dados centrais.
- **Application Layer**: Implementa a lógica de negócios através de manipuladores de comandos e consultas.
- **Infrastructure Layer**: Fornece implementações concretas de repositórios e buses para comandos, consultas e eventos.

#### Integração com Watermill

O Watermill é utilizado para gerenciar a comunicação de mensagens entre diferentes partes do sistema. Ele suporta diferentes adaptadores de *message broker*, permitindo que o sistema seja facilmente escalado ou distribuído em diferentes serviços.

### Considerações Finais

Este projeto serve como uma base sólida para implementar sistemas orientados a mensagens em Go. Ele demonstra boas práticas de arquitetura, como CQRS e uso de *message brokers*, que podem ser aplicadas a outros contextos de negócios ou ampliadas para incluir funcionalidades adicionais, como autenticação, autorização, e persistência em banco de dados. A capacidade de trocar facilmente o *message broker* subjacente permite que o sistema se adapte a diferentes requisitos de carga e distribuição, tornando-o uma solução flexível e escalável para sistemas modernos.
