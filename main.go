package main

import (
	"embed"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//go:embed templates/*
var templateFS embed.FS

type Card struct {
	Suit   string
	Value  string
	Points int
}

type Player struct {
	Balance  float64
	Bet      float64
	Hand     []Card
	Score    int
	Standing bool
}

type Game struct {
	Player   Player
	Dealer   Player
	Deck     []Card
	GameOver bool
	Message  string
	mu       sync.Mutex
}

var (
	game      *Game
	templates *template.Template
)

func initGame() *Game {
	return &Game{
		Player: Player{Balance: 1000.0, Bet: 0.0},
		Dealer: Player{},
		Deck:   shuffleDeck(createDeck()),
	}
}

func createDeck() []Card {
	suits := []string{"♥", "♦", "♠", "♣"}
	values := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	var deck []Card

	for _, suit := range suits {
		for _, value := range values {
			points := 0
			switch value {
			case "A":
				points = 11
			case "K", "Q", "J":
				points = 10
			default:
				points, _ = strconv.Atoi(value)
			}
			deck = append(deck, Card{Suit: suit, Value: value, Points: points})
		}
	}
	return deck
}

func shuffleDeck(deck []Card) []Card {
	rand.Seed(time.Now().UnixNano())
	for i := len(deck) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}
	return deck
}

func calculateScore(hand []Card) int {
	score := 0
	aces := 0

	for _, card := range hand {
		if card.Value == "A" {
			aces++
		}
		score += card.Points
	}

	for score > 21 && aces > 0 {
		score -= 10
		aces--
	}

	return score
}

func dealCard() Card {
	if len(game.Deck) == 0 {
		game.Deck = shuffleDeck(createDeck())
	}
	card := game.Deck[0]
	game.Deck = game.Deck[1:]
	return card
}

func startNewRound() {
	// Deal initial cards
	game.Player.Hand = []Card{dealCard(), dealCard()}
	game.Dealer.Hand = []Card{dealCard(), dealCard()}

	// Calculate initial scores
	game.Player.Score = calculateScore(game.Player.Hand)
	game.Dealer.Score = calculateScore(game.Dealer.Hand)

	// Reset game state
	game.GameOver = false
	game.Player.Standing = false
	game.Message = "Ihre Runde! Hit oder Stand?"

	// Check for initial blackjack
	if game.Player.Score == 21 {
		handleDealerTurn()
		determineWinner()
	}
}

func handleDealerTurn() {
	for calculateScore(game.Dealer.Hand) < 17 {
		game.Dealer.Hand = append(game.Dealer.Hand, dealCard())
	}
	game.Dealer.Score = calculateScore(game.Dealer.Hand)
}

func determineWinner() {
	playerScore := game.Player.Score
	dealerScore := game.Dealer.Score

	// Check for blackjack (21 with exactly 2 cards)
	playerBlackjack := playerScore == 21 && len(game.Player.Hand) == 2
	dealerBlackjack := dealerScore == 21 && len(game.Dealer.Hand) == 2

	if playerScore > 21 {
		game.Player.Balance -= game.Player.Bet
		game.Message = "Bust! Sie haben verloren! (" + strconv.Itoa(playerScore) + " Punkte)"
	} else if dealerScore > 21 {
		game.Player.Balance += game.Player.Bet
		game.Message = "Dealer bust! Sie haben gewonnen! (" + strconv.Itoa(dealerScore) + " Punkte)"
	} else if playerBlackjack && !dealerBlackjack {
		// Blackjack pays 3:2
		game.Player.Balance += game.Player.Bet * 1.5
		game.Message = "Blackjack! Sie haben gewonnen! (1.5x Gewinn)"
	} else if !playerBlackjack && dealerBlackjack {
		game.Player.Balance -= game.Player.Bet
		game.Message = "Dealer hat Blackjack! Sie haben verloren!"
	} else if playerScore > dealerScore {
		game.Player.Balance += game.Player.Bet
		game.Message = "Sie haben gewonnen! (" + strconv.Itoa(playerScore) + " vs " + strconv.Itoa(dealerScore) + ")"
	} else if dealerScore > playerScore {
		game.Player.Balance -= game.Player.Bet
		game.Message = "Dealer gewinnt! (" + strconv.Itoa(dealerScore) + " vs " + strconv.Itoa(playerScore) + ")"
	} else {
		game.Message = "Unentschieden! (" + strconv.Itoa(playerScore) + " Punkte)"
	}

	game.GameOver = true
	game.Player.Bet = 0.0
}

func handleRestartRound(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	game.mu.Lock()
	game.Player.Standing = false
	game.Player.Bet = 0.0
	game.Player.Hand = nil
	game.Dealer.Hand = nil
	game.GameOver = false
	game.Message = ""
	game.mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	game = initGame()

	templates = template.Must(template.ParseFS(templateFS, "templates/*.html"))

	http.HandleFunc("/", handleGame)
	http.HandleFunc("/placeBet", handlePlaceBet)
	http.HandleFunc("/hit", handleHit)
	http.HandleFunc("/stand", handleStand)
	http.HandleFunc("/newGame", handleNewGame)
	http.HandleFunc("/restartRound", handleRestartRound)

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleGame(w http.ResponseWriter, r *http.Request) {
	game.mu.Lock()
	defer game.mu.Unlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "game.html", game); err != nil {
		log.Printf("Template execution error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}
}

func handlePlaceBet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bet, err := strconv.ParseFloat(r.FormValue("bet"), 64)
	if err != nil || bet <= 0.0 || bet > game.Player.Balance {
		http.Error(w, "Invalid bet amount", http.StatusBadRequest)
		return
	}

	game.mu.Lock()
	game.Player.Bet = bet
	startNewRound()
	game.mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleHit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	if game.GameOver || game.Player.Standing {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Deal a new card to the player
	newCard := dealCard()
	game.Player.Hand = append(game.Player.Hand, newCard)

	// Update player's score
	game.Player.Score = calculateScore(game.Player.Hand)

	// Check for bust
	if game.Player.Score > 21 {
		handleDealerTurn() // Let dealer take their turn even if player busts
		determineWinner()
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleStand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	if game.GameOver {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	game.Player.Standing = true
	handleDealerTurn()
	determineWinner()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleNewGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	game.mu.Lock()
	if game.Player.Balance <= 0.0 {
		game.Player.Balance = 1000.0 // Give player more money if they're broke
		game.Message = "Willkommen zurück! Sie haben €1000 Startguthaben erhalten."
	}
	game.Player.Standing = false
	game.Player.Bet = 0.0
	game.GameOver = false
	game.Player.Hand = nil
	game.Dealer.Hand = nil
	game.mu.Unlock()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
