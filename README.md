# e2e-api-featureflags

Ein Feature-Flag-Service als REST-API in Go. Der Dienst verwaltet Feature-Flags
in einem thread-sicheren In-Memory-Store und stellt sie über REST-Endpunkte
bereit: anlegen, listen, lesen, ändern, löschen sowie ein Evaluate-Endpunkt für
eine deterministische Ja/Nein-Entscheidung pro Nutzer anhand des
`rollout_percent`.

## Tech Stack

- **Sprache**: Go (1.22+)
- **Framework**: `net/http` (Standardbibliothek, keine externen Abhängigkeiten)
- **Store**: In-Memory mit `sync.RWMutex`
- **Tests**: `go test` + `net/http/httptest`

## Installation

Es sind keine externen Abhängigkeiten nötig. Es genügt Go 1.22 oder neuer.

```sh
go mod download
```

## Start (Development)

```sh
go run .
```

Der Server lauscht anschließend auf `http://localhost:8080`.

## Build (Production)

```sh
go build ./...
```

Das erzeugt ein einzelnes Binary `e2e-api-featureflags`, das auf `:8080` lauscht.

## Endpunkte

| Methode | Pfad                          | Beschreibung                                        | Erfolg |
|---------|-------------------------------|-----------------------------------------------------|--------|
| POST    | `/flags`                      | Legt ein Flag an                                    | 201    |
| GET     | `/flags`                      | Listet alle Flags (leerer Store → `[]`)             | 200    |
| GET     | `/flags/{key}`                | Liefert ein einzelnes Flag                          | 200    |
| PUT     | `/flags/{key}`                | Ändert enabled/description/rollout_percent          | 200    |
| DELETE  | `/flags/{key}`                | Löscht ein Flag                                     | 204    |
| GET     | `/flags/{key}/evaluate?user=` | Deterministische Rollout-Entscheidung `{"result":bool}` | 200 |
| GET     | `/healthz`                    | Health-Check                                        | 200    |

Fehlerantworten haben immer die Form `{"error":"..."}` mit passendem Statuscode
(400 Validierung, 404 unbekannt, 405 Methode, 409 Duplikat) und enthalten keine
internen Details oder Stacktraces. Antworten setzen keine
`Access-Control-Allow-Origin`-Header.

### Flag-JSON

```json
{
  "key": "my-flag",
  "enabled": true,
  "description": "…",
  "rollout_percent": 50
}
```

`rollout_percent` liegt im Bereich 0–100 (Default 0).

## Tests

```sh
go test ./...
go test -race ./...
```

## Features

- Thread-sicherer In-Memory-Store (`sync.RWMutex` + `map[string]Flag`)
- CRUD über REST-Endpunkte
- Deterministischer Evaluate-Endpunkt per stabilem Hash
- Einheitliche JSON-Fehlerobjekte ohne interne Details
- Request-Logging-Middleware (Methode, Pfad ohne Query, Statuscode)
