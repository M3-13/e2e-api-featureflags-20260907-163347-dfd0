VERDICT: APPROVED

## Sicherheitsbericht

Der vollständige Stand des Feature-Flag-Service wurde anhand der im Sprint-Spec definierten Sicherheitskriterien (AC-15 bis AC-20) geprüft. Es wurden keine exploitable Sicherheitslücken festgestellt, die eines der Sicherheits-Acceptance-Kriterien verletzen.

### Erfüllte Sicherheitskriterien

- **AC-15 – Request-Body-Begrenzung:**  
  `respond.go` begrenzt eingehende JSON-Bodies über `http.MaxBytesReader` auf `1 << 20` Bytes (1 MB). Die Funktion `decodeJSON` wird sowohl von `postFlags` als auch von `putFlag` verwendet. Damit ist die Anforderung erfüllt.

- **AC-16 / AC-20 – Saubere Fehlerantworten:**  
  Alle Fehlerantworten werden über `writeError` erzeugt und enthalten ausschließlich das JSON-Feld `error`. Es werden statische, nicht-technische Meldungen ausgegeben. Stacktraces, Dateipfade oder interne Fehlermeldungen werden nicht preisgegeben.

- **AC-17 – Maximallänge des `user`-Parameters:**  
  `handlers_evaluate.go` prüft `len(user) > 255` und antwortet bei Überschreitung mit `400 Bad Request`. Überlange Werte werden damit abgewiesen.

- **AC-18 – Keine CORS-Header:**  
  Es werden an keiner Stelle `Access-Control-Allow-Origin`-Header gesetzt, weder explizit noch durch Middleware. Browserbasierte Cross-Origin-Aufrufe werden nicht aktiv unterstützt.

- **AC-19 – Datensparsames Logging:**  
  Die Logging-Middleware `logRequests` protokolliert ausschließlich HTTP-Methode, `r.URL.Path` (ohne Query-String) und den Statuscode. Der `user`-Query-Parameter ist nie Bestandteil der Logzeile. Der zugehörige Test deckt dies explizit ab.

### Prüfbereiche ohne Befund

- **Secrets:**  
  Keine hartkodierten Schlüssel, Passwörter, Token oder sensiblen URLs im Code gefunden. Logging erfolgt ohne sensitive Daten.

- **Injection / Input-Validierung:**  
  Es gibt keine SQL-, Command- oder Path-Injection. Path-Parameter (`r.PathValue("key")`) werden als Schlüssel in einer In-Memory-Map verwendet, nicht als Datei- oder Shell-Pfad. JSON-Deserialisierung erfolgt mit der Standardbibliothek ohne unsichere Typen oder Deserialisierungs-Primitiven. XSS ist wegen `Content-Type: application/json` und fehlender HTML-Ausgabe nicht relevant.

- **AuthN/AuthZ:**  
  Es gibt keine Authentifizierung oder Autorisierung. Dies ist im Sprint-Spec jedoch nicht als Kriterium gefordert und wird daher nicht als Verstoß gewertet. Siehe Hinweise.

- **Dependencies / Scanner:**  
  Das Projekt nutzt ausschließlich die Go-Standardbibliothek, es gibt keine externen Abhängigkeiten. Der Scanner-Abschnitt meldet „no applicable security scanners for this project type“, daher ist kein automatisches Scan-Ergebnis vorhanden. Mangels Ausgabe wird hieraus kein Befund abgeleitet.

- **Transport / Konfiguration:**  
  Der Dienst lauscht unverschlüsselt auf `:8080` (`http.ListenAndServe`). Ein TLS-Abschluss ist nicht Teil der Anforderungen. Die Logging-Middleware gibt Logs unmittelbar per `fmt.Printf` auf `stdout` aus; eine Log-Injection über manipulierte Pfade ist nicht möglich, da `r.URL.Path` keine Steuerzeichen aus dem Request enthält.

### Hinweise (non-blocking)

- **Fehlende Authentifizierung / Autorisierung:**  
  Der Service ist vollständig offen – jede erreichbare Instanz kann Flags anlegen, ändern und auswerten. Für einen produktiven Einsatz oder eine Exposition außerhalb vertrauenswürdiger Netze ist eine Authentifizierung dringend zu empfehlen. Da der Sprint-Spec dies nicht als Kriterium aufführt, bleibt es eine Notiz für die nächste Planungsrunde.

- **Fehlendes TLS:**  
  Der Server verwendet `http.ListenAndServe` ohne TLS. Vertrauliche Daten (z. B. Flag-Konfiguration oder Evaluate-Ergebnisse) würden im Klartext übertragen. Ebenfalls nicht Teil des aktuellen Specs.

- **Fehlendes Rate-Limiting:**  
  Es existiert keine Begrenzung der Anfragemenge. In Kombination mit der fehlenden Auth kann der Service leicht als Ziel für Ressourcen-Erschöpfung dienen. Nicht in den ACs gefordert.

- **Unterschiedliche Behandlung übergroßer PUT-Bodies:**  
  `putFlag` übersetzt einen `http.MaxBytesError` aus `decodeJSON` in ein pauschales `400 invalid JSON body`, während `postFlags` hier korrekt `413 Request Entity Too Large` liefert. Die Begrenzung selbst ist vorhanden (AC-15 erfüllt), eine Unterscheidung wäre für sauberes API-Verhalten jedoch konsistenter.

- **Längenprüfung in Bytes statt Unicode-Zeichen:**  
  `len(user)` zählt Bytes, nicht Unicode-Zeichen. Bei mehrbyte-UTF-8-Zeichen kann ein Wert mit weniger als 255 Zeichen bereits abgelehnt werden. Funktional streng, aber kein Sicherheitsrisiko und im getesteten ASCII-Fall konform. Für eine exakte „Zeichen“-Semantik wäre `utf8.RuneCountInString` zu prüfen.

### Ergebnis

Alle im Sprint-Spec definierten Sicherheits-Acceptance-Kriterien (AC-15 bis AC-20) werden durch den vorliegenden Code erfüllt. Es gibt keine blockierenden oder änderungswürdigen Sicherheitsbefunde im Sinne der Vorgaben. Die genannten Hinweise sind Empfehlungen für die nächste Planungsiteration und tragen nicht zur aktuellen Verdict-Entscheidung bei.