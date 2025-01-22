# Blackjack Game

Ein einfaches Blackjack-Spiel, implementiert in Go mit HTML-Templates.

## Features

- Modernes und ansprechendes Web-Interface
- Geldkonto-System mit Einsatz-Management
- Vollständige Blackjack-Spiellogik
- Responsive Design

## Installation

### Option 1: Vorkompilierte Binaries

Laden Sie die neueste Version für Ihr Betriebssystem von der [Releases-Seite](https://github.com/aarxns/goblackjack/releases) herunter:

- Windows: `blackjack-windows-amd64.zip`
- macOS (Apple Silicon): `blackjack-darwin-arm64.zip`
- Linux: `blackjack-linux-amd64.zip`

Entpacken Sie die ZIP-Datei und führen Sie die Binary aus.

### Option 2: Aus dem Quellcode

Voraussetzungen:
- Go 1.21 oder höher

1. Klonen Sie das Repository:
```bash
git clone https://github.com/aarxns/goblackjack.git
cd goblackjack
```

2. Starten Sie den Server:
```bash
go run main.go
```

## Spielstart

Öffnen Sie einen Webbrowser und navigieren Sie zu:
```
http://localhost:8080
```

## Spielregeln

- Sie starten mit einem Guthaben von €1000
- Setzen Sie einen Betrag vor jeder Runde
- Versuchen Sie, näher an 21 Punkte zu kommen als der Dealer
- Ass zählt als 1 oder 11 Punkte
- Bildkarten (J, Q, K) zählen als 10 Punkte
- Überschreiten Sie 21 Punkte, verlieren Sie automatisch

## Spielablauf

1. Setzen Sie Ihren Einsatz
2. Entscheiden Sie sich für "Hit" (eine weitere Karte) oder "Stand" (keine weiteren Karten)
3. Der Dealer zieht Karten, bis er mindestens 17 Punkte hat
4. Der Gewinner wird ermittelt und der Einsatz entsprechend verteilt
