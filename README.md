# Chat App Backend (Go, WebSocket, PostgreSQL)

Backend chat realtime dengan pendekatan Clean Architecture yang dirancang agar mudah di-maintain, mudah diuji, dan siap di-scale dari MVP ke production.

## 1. Gambaran Arsitektur Sistem

### Tech Stack

- Backend: Go
- Realtime Transport: WebSocket (Gorilla WebSocket)
- Database: PostgreSQL (pgx)
- File Storage: Local storage (`/uploads`) dengan opsi migrasi ke S3 compatible (MinIO, AWS S3)
- Auth: JWT

### High-Level Architecture

```text
Client (Web / Mobile)
        |
        | WebSocket + REST
        v
API Gateway / HTTP Server
        |
        v
Application Layer (Usecase)
        |
        v
Domain Layer (Entities)
        |
        v
Repository Layer
        |
        v
PostgreSQL
```

### Realtime Flow

```text
User A -> WebSocket -> Server -> Broadcast -> User B
```

## 2. Fitur Utama

### 2.1 Realtime Messaging

- Kirim pesan ke room
- Terima pesan realtime
- Simpan histori pesan ke PostgreSQL
- Ambil histori via REST (pagination)

### 2.2 Online User Indicator

- User connect -> emit `user_online`
- User disconnect -> emit `user_offline`
- Broadcast status ke client lain

### 2.3 Typing Indicator

Event:

- `typing_start`
- `typing_stop`

Flow:

```text
User A mengetik
-> send typing event
-> server broadcast ke user lain di room
-> UI menampilkan "User A sedang mengetik..."
```

### 2.4 Chatroom

Tipe room:

- Private chat
- Group chat

Tabel inti:

- `rooms`
- `room_members`
- `messages`

### 2.5 Kirim File / Gambar

Flow:

```text
Client upload file -> API /upload
Server simpan ke storage
Server kirim file URL lewat event websocket
```

Contoh payload pesan file:

```json
{
  "type": "image",
  "url": "/uploads/img123.png"
}
```

## 3. Struktur Project (Clean Architecture)

Struktur saat ini:

```text
chat-app
|
|-- cmd/
|   `-- server/
|       `-- main.go
|
|-- internal/
|   |-- delivery/
|   |   `-- http/
|   |       |-- auth_handler.go
|   |       |-- message_handler.go
|   |       `-- upload_handler.go
|   |
|   |-- domain/                  # entity domain (saat ini masih kosong)
|   |-- infrastructure/
|   |   |-- database/
|   |   |   `-- postgres.go
|   |   `-- storage/
|   |
|   |-- repository/
|   |   |-- auth_repository.go
|   |   `-- message_repository.go
|   |
|   |-- usecase/
|   |   `-- message_usecase.go
|   |
|   `-- websocket/
|       |-- client.go
|       |-- handler.go
|       |-- hub.go
|       `-- message.go
|
|-- migrations/
|-- pkg/
|   `-- utils/
|       `-- jwt.go
|-- uploads/
|-- go.mod
`-- go.sum
```

### Mapping Layer

- Delivery layer: parsing HTTP/WebSocket request + response
- Usecase layer: business logic aplikasi
- Domain layer: entity/value object/interface contract
- Repository layer: akses data PostgreSQL
- Infrastructure layer: detail teknis DB/storage

## 4. Domain Entity (Blueprint)

Berikut blueprint entity yang direkomendasikan untuk domain layer:

```go
type User struct {
    ID        string
    Username  string
    Online    bool
    CreatedAt time.Time
}

type Room struct {
    ID        string
    Name      string
    Type      string // private | group
    CreatedBy string
    CreatedAt time.Time
}

type Message struct {
    ID        string
    SenderID  string
    RoomID    string
    Content   string
    FileURL   string
    Type      string // text | image | file
    Status    string // sent | delivered | read
    CreatedAt time.Time
}
```

## 5. WebSocket Hub (Core Realtime Engine)

Pattern utama yang dipakai:

- Register client
- Unregister client
- Broadcast event per room
- Persist message saat event `send_message`

Blueprint struktur hub:

```go
type Hub struct {
    clients    map[*Client]bool
    rooms      map[string]map[*Client]bool
    register   chan *Client
    unregister chan *Client
    broadcast  chan *RoomMessage
}
```

Loop inti (ringkas):

```go
for {
    select {
    case client := <-h.register:
        h.clients[client] = true
    case client := <-h.unregister:
        delete(h.clients, client)
    case msg := <-h.broadcast:
        for c := range h.rooms[msg.roomID] {
            c.send <- msg.raw
        }
    }
}
```

## 6. Database Design

Skema konseptual:

### users

- id (uuid, pk)
- username (unique)
- password
- created_at

### rooms

- id (uuid, pk)
- name
- type (`private` | `group`)
- created_by (fk -> users.id)
- created_at

### room_members

- user_id (fk -> users.id)
- room_id (fk -> rooms.id)
- joined_at
- primary key (user_id, room_id)

### messages

- id (uuid, pk)
- room_id (fk -> rooms.id)
- sender_id (fk -> users.id)
- content
- file_url
- type (`text` | `image` | `file`)
- status (`sent` | `delivered` | `read`)
- created_at

Contoh SQL inisialisasi (karena folder migration masih kosong):

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'group',
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS room_members (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, room_id)
);

CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT,
    file_url TEXT,
    type VARCHAR(20) NOT NULL DEFAULT 'text',
    status VARCHAR(20) NOT NULL DEFAULT 'sent',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_room_created_at
ON messages(room_id, created_at);
```

## 7. Event System (WebSocket)

Gunakan event-based message dengan envelope standar:

```json
{
  "event": "send_message",
  "data": {
    "room_id": "123",
    "message": "Hello"
  }
}
```

Event yang dipakai/direncanakan:

- `join_room`
- `leave_room` (blueprint, belum aktif)
- `send_message`
- `receive_message`
- `typing_start`
- `typing_stop`
- `user_online`
- `user_offline`
- `message_delivered`
- `message_read`

## 8. Flow Kirim Pesan

```text
Client
   |
   | send_message
   v
WebSocket Handler
   |
   v
Usecase SendMessage
   |
   v
Save to PostgreSQL
   |
   v
Hub Broadcast
   |
   v
Clients receive_message
```

## 9. API dan Endpoint Saat Ini

### Auth

- `POST /register`
- `POST /login`

### Realtime

- `GET /ws?token=<jwt>`

### Message History

- `GET /messages?room_id=<id>&page=1&limit=20`

### Upload

- `POST /upload` (multipart form-data, key: `file`)
- Static file: `GET /uploads/<filename>`

## 10. Setup Lokal

### Prasyarat

- Go (disarankan 1.22+)
- PostgreSQL 14+

### Environment Variable

Buat file `.env` di root project:

```env
DATABASE_URL=postgres://postgres:password@localhost:5432/chat_app?sslmode=disable
BASE_URL=http://localhost:8080
```

### Jalankan Aplikasi

```bash
go mod tidy
go run ./cmd/server
```

Server default berjalan di:

- `http://localhost:8080`

## 11. Rekomendasi Library

- WebSocket: `github.com/gorilla/websocket`
- PostgreSQL: `github.com/jackc/pgx/v5`
- Migration: `golang-migrate` (direkomendasikan ditambahkan)
- UUID: `github.com/google/uuid`
- JWT: `github.com/golang-jwt/jwt/v5`

## 12. Blueprint Scalability (Menuju Production)

### 12.1 Presence System

- Simpan online/offline state ke Redis (opsional)
- Simpan last_seen ke PostgreSQL

### 12.2 Delivery Guarantee

- Track status `sent -> delivered -> read`
- Retry event yang gagal acknowledge

### 12.3 Pagination dan Query Performance

- Gunakan keyset pagination untuk room besar
- Tambah index komposit sesuai pola query

### 12.4 Rate Limiting

- Batasi event `send_message` per user
- Lindungi endpoint upload dari abuse

### 12.5 Horizontal Scaling

- Pisahkan WebSocket gateway dari service logic
- Gunakan pub/sub (Redis/NATS/Kafka) untuk cross-instance broadcast

### 12.6 Storage Abstraction

- Definisikan interface storage di `internal/infrastructure/storage`
- Implementasi local + S3 compatible tanpa ubah usecase

## 13. Roadmap Implementasi

### Phase 1 (MVP)

- Auth login/register
- Join room + send/receive message
- Message history
- Upload file/image

### Phase 2

- Typing indicator stabil
- Online/offline presence stabil
- Delivery/read receipt penuh

### Phase 3

- Group management (invite/kick/role)
- Redis pub/sub
- Observability (structured logging, metrics, tracing)

## 14. Target Skill yang Didapat

Dengan menyelesaikan project ini, kamu akan melatih:

- Realtime system design
- WebSocket architecture
- Clean architecture di Go
- Database modeling untuk chat domain
- Concurrent programming di Go
- Scalable backend engineering

---

Jika kamu ingin, tahap berikutnya bisa langsung dibuatkan:

- Skeleton domain entity + interface repository
- Migrasi SQL awal ke folder `migrations`
- Usecase `SendMessage`, `JoinRoom`, `Typing`
- Refactor handler agar sepenuhnya lewat usecase layer
