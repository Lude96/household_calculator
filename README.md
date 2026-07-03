# Household Calculator

Ein CLI-Tool zur Verwaltung von Haushaltsausgaben und Einnahmen mit automatischer Grafik- und PDF-Generierung.

## Features

**Funktionalität:**
- Ausgaben und Einnahmen erfassen
- Grafische Visualisierung der Ausgaben (HTML Chart)
- PDF-Reports generieren
- Transaktionen auflisten und verwalten
- Kategorisierung von Transaktionen
- Persistente Speicherung in JSON

## Installation & Verwendung

### Mit Docker

```bash
docker build -t household-calculator .

docker run -v $(pwd)/data:/app/data household-calculator add

docker run -v $(pwd)/data:/app/data household-calculator add \
  -type expense -category Groceries -amount 50.25 -notes "Wochenmarkt"

docker run -v $(pwd)/data:/app/data household-calculator list

docker run -v $(pwd)/data:/app/data household-calculator report -format pdf
docker run -v $(pwd)/data:/app/data household-calculator report -format chart
```

### Lokal (ohne Docker)

```bash
go mod tidy
go build -o household_calculator

./household_calculator add
./household_calculator add -type expense -category Groceries -amount 50.25
./household_calculator list
./household_calculator report -format pdf
./household_calculator report -format chart
```

## Befehle

### add - Transaktion hinzufügen

```bash
household add [options]

Options:
  -type      income|expense (erforderlich)
  -category  Kategorie-Name (erforderlich)
  -amount    Betrag (erforderlich)
  -notes     Optionale Notizen
```

**Beispiele:**
```bash
household add -type income -category Salary -amount 3000
household add -type expense -category Groceries -amount 45.50 -notes "Wochenmarkt"
household add -type expense -category Transport -amount 25
```

### list - Alle Transaktionen anzeigen

```bash
household list
```

Zeigt:
- Alle Transaktionen mit Kategorie und Betrag
- Gesamteinnahmen
- Gesamtausgaben
- Bilanz (Überschuss/Defizit)

### report - Report generieren

```bash
household report -format [console|pdf|chart]
```

**Formate:**
- `console` (default): Textanzeige mit Zusammenfassung
- `pdf`: Generiert `report.pdf` mit allen Details
- `chart`: Generiert `expenses_chart.html` mit interaktiver Grafik

### clear - Daten löschen

```bash
household clear
```

Löscht alle Transaktionen und setzt das System zurück.

## Datenspeicherung

Alle Daten werden in `household_data.json` gespeichert. Das Format:

```json
{
  "transactions": [
    {
      "category": "Groceries",
      "amount": 45.5,
      "type": "expense",
      "notes": "Wochenmarkt"
    }
  ]
}
```

## Ausgabeformate

### Text-Report (console)
```
Total Income: 3000.00
Total Expenses: 70.50
Balance: 2929.50
```

### PDF-Report
- Zusammenfassung (Gesamteinnahmen, Ausgaben, Bilanz)
- Ausgaben nach Kategorien
- Detaillierte Transaktionsliste

### HTML-Grafik (chart)
- Interaktives Kreisdiagramm (Pie Chart)
- Visualisierung der Ausgabenverteilung
- Balance als "Überschuss" Kategorie

## Technologie

- **Sprache:** Go 1.26
- **Chart-Library:** go-echarts
- **PDF-Library:** gofpdf
- **Container:** Docker (Alpine Linux)

