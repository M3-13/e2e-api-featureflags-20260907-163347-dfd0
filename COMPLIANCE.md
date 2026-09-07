VERDICT: APPROVED

## Prüfbericht

### 1. DSGVO / Datenschutz
**Befund:** Der Query-Parameter `user` ist potenziell personenbezogen, wird aber ausschließlich transient für die Hash-Berechnung verwendet. Er wird weder gespeichert noch geloggt. Die Pflichtkriterien sind erfüllt:
- **AC-19** erfüllt: `middleware.go` protokolliert ausschließlich `r.Method`, `r.URL.Path` (ohne Query-String) und Statuscode. Der `user`-Parameter erscheint nicht in den Logs.
- **AC-20** erfüllt: Fehlerantworten über `writeError` enthalten ausschließlich das JSON-Feld `error`; keine Stacktraces, Dateipfade oder internen Fehlermeldungen.
- **AC-16** erfüllt: Es werden keine internen Details in Fehlerantworten ausgegeben.
- **AC-17** erfüllt: `user` wird auf maximal 255 Zeichen begrenzt, längere Werte führen zu 400 Bad Request.

**Notes (non-blocking):**
- Der Dienst lauscht auf `:8080` ohne TLS (`main.go`). Da der `user`-Parameter personenbezogen sein kann, wäre eine TLS-Absicherung der Übertragung datenschutzrechtlich geboten. Dies ist nicht durch ein AC-Kriterium abgedeckt und daher als Hinweis für den nächsten Planungspass zu verstehen.
- Flag-Keys und `description` werden als Teil des Pfades (`/flags/{key}`) geloggt. Sofern Betreiber dort personenbezogene Daten ablegen, entstünde ein Logging-Risiko. Die Datenminimierung sollte in der Betreiberdokumentation geregelt werden.

### 2. EU Cyber Resilience Act (CRA)
**Befund:** Grundlegende Sicherheitsanforderungen sind umgesetzt:
- **AC-15** erfüllt: Request-Bodies werden über `http.MaxBytesReader` auf 1 MB begrenzt (`respond.go`).
- **AC-16 / AC-20** erfüllt: Keine internen Fehlerdetails in Antworten.
- **AC-18** erfüllt: Keine `Access-Control-Allow-Origin`-Header.
- Der Store ist thread-sicher (`sync.RWMutex`), die Race-Tests sind vorhanden.

**Notes (non-blocking):**
- Es fehlt ein dokumentiertes SBOM bzw. eine vollständige Abhängigkeits-/Lizenzliste über `go.mod` hinaus. Die Abhängigkeiten sind hier minimal (Standardbibliothek), aber für ein Inverkehrbringen nach CRA sollte ein SBOM erzeugt werden.
- Es ist kein dokumentierter Update-/Patch-Prozess sichtbar; als Quellcode-Backend kann dies über den Build-/Deployment-Prozess des Betreibers erfolgen.
- Keine Authentifizierung/Autorisierung/Rate-Limiting. Dies ist nicht durch ein AC-Kriterium abgedeckt, aber als Security-by-design-Lücke für den nächsten Planungspass relevant.

### 3. EU KI-Verordnung (AI Act)
Nicht anwendbar — es ist kein KI-Feature im Produkt sichtbar.

### 4. Pflichttexte & UI
Nicht anwendbar — reiner Backend-Dienst ohne Endnutzer-UI. Es bestehen keine Impressums-, Datenschutzerklärungs- oder Cookie-Pflichten in diesem Artefakt.

### 5. Barrierefreiheit
Nicht anwendbar — keine öffentliche Web-UI.

### Sonstige nicht-blockierende Beobachtung
- `handlers_update.go` behandelt alle `decodeJSON`-Fehler pauschal als 400 „invalid JSON body“, auch wenn ein `http.MaxBytesError` vorliegt. Empfehlung: analog zu `handlers_create.go` bei Überschreiten der Größenbegrenzung 413 „request body too large“ zurückgeben, um konsistente Statuscodes zu gewährleisten. Kein AC-Verstoß.