package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

const Back = "1B"
const Blank = "2B" // Used for empty discard pile
const Wrong = "X"  // Used for incorrect tag
const StartHandSize = 4

// Game states
const (
	StartGame string = "startGame"
	Playing   string = "playing"
	EndGame   string = "endGame"
)

// Misc
const (
	Final   string = "final"
	Ready   string = "ready"
	Draw    string = "draw"
	Discard string = "discard"
	Card    string = "card"
	Player  string = "player"
	Owner   string = "owner"
	Index   string = "index"
	Bunga   string = "bunga"
)

// Playing states
const (
	StartTurn          string = "startTurn"
	DiscardSwapChoice  string = "discardSwapChoice"
	DrawChoice         string = "drawChoice"
	LookOwnChoice      string = "lookOwnChoice"
	LookingOwn         string = "lookingOwn"
	LookOtherChoice    string = "lookOtherChoice"
	LookingOther       string = "lookingOther"
	SwapOtherChoice    string = "swapOtherChoice"
	SwapOtherOwnChoice string = "swapOtherOwnChoice"
	LookSwapChoice     string = "lookSwapChoice"
	LookSwapOwnChoice  string = "lookSwapOwnChoice"
)

// Teams
const (
	Players    string = "Players"
	Spectators string = "Spectators"
)

// Highlight letters for cards
const (
	PrimHl string = "p"
	SecoHl string = "s"
	BungHl string = "b"
)

type bungaAction struct {
	Start    string
	StartIdx string
	End      string
	EndIdx   string
	Card     string
}

type bungaUserState struct {
	DrawPile     string
	DiscardPile  string
	LatestAction []bungaAction
	Turn         string
	PlayersReady map[string]string
	PlayerHands  map[string][]string
	SaidBunga    string
	Scores       map[string]int
	PlayerOrder  []string
	PlayingState string
	Winner       string
}

type bungaGameState struct {
	DrawPile     []string
	DiscardPile  []string
	LatestAction []bungaAction
	LatestTag    string
	CardSelected string
	Turn         string
	PlayersReady map[string]string
	PlayerHands  map[string][]string
	SaidBunga    string
	GameState    string
	PlayingState string
	Scores       map[string]int
	PlayerOrder  []string
	Winner       string
}

type bunga struct {
	in    chan userMsg
	out   chan gameMsg
	lobby *lobbyState
	state bungaGameState
}

func (b *bunga) reshuffleDiscardPile() {
	// put all but the top card of the discard pile back into the draw pile, then shuffle the draw pile
	for i := 0; i < len(b.state.DiscardPile)-1; i++ {
		b.state.DrawPile = append(b.state.DrawPile, b.state.DiscardPile[i])
	}
	rand.Shuffle(len(b.state.DrawPile), func(i, j int) {
		b.state.DrawPile[i], b.state.DrawPile[j] = b.state.DrawPile[j], b.state.DrawPile[i]
	})
}

func (b *bunga) drawCard() string {
	card := b.drawTop()
	b.state.DrawPile = b.state.DrawPile[:len(b.state.DrawPile)-1]
	if len(b.state.DrawPile) == 1 {
		b.reshuffleDiscardPile()
	}
	return card
}

func (b *bunga) drawTop() string {
	return b.state.DrawPile[len(b.state.DrawPile)-1]
}

func (b *bunga) discardTop() string {
	if len(b.state.DiscardPile) == 0 {
		return Blank
	}
	return b.state.DiscardPile[len(b.state.DiscardPile)-1]
}

func (b *bunga) advanceTurn() {
	curTurnIdx := 0
	for i, player := range b.state.PlayerOrder {
		if b.state.Turn == player {
			curTurnIdx = i
		}
	}
	nextTurnIdx := (curTurnIdx + 1) % len(b.state.PlayerOrder)
	b.state.Turn = b.state.PlayerOrder[nextTurnIdx]
	if b.state.Turn == b.state.SaidBunga {
		b.state.GameState = EndGame
	}
}

func (b *bunga) initDeckCards() {
	suits := []string{"C", "D", "H", "S"}
	names := []string{"2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K", "A"}
	b.state.DrawPile = []string{}
	for _, suit := range suits {
		for _, name := range names {
			b.state.DrawPile = append(b.state.DrawPile, name+suit)
		}
	}
	rand.Shuffle(len(b.state.DrawPile), func(i, j int) {
		b.state.DrawPile[i], b.state.DrawPile[j] = b.state.DrawPile[j], b.state.DrawPile[i]
	})
}

func (b *bunga) initPlayerHands() {
	b.state.PlayerHands = map[string][]string{}
	for _, player := range b.state.PlayerOrder {
		for i := 0; i < StartHandSize; i++ {
			b.state.PlayerHands[player] = append(b.state.PlayerHands[player], b.drawCard())
		}
	}
}

func (b *bunga) initPlayerOrder() {
	order := []string{}
	for player, _ := range b.lobby.Players {
		order = append(order, player)
	}
	rand.Shuffle(len(order), func(i, j int) {
		order[i], order[j] = order[j], order[i]
	})
	b.state.PlayerOrder = order
}

func createBunga(l *lobbyState, in chan userMsg, out chan gameMsg) game {
	ret := &bunga{
		in:    in,
		out:   out,
		lobby: l,
		state: bungaGameState{
			DiscardPile: []string{},
			SaidBunga:   "",
			GameState:   StartGame,
		},
	}
	ret.initDeckCards()
	ret.initPlayerOrder()
	ret.state.Turn = ret.state.PlayerOrder[0]
	ret.initPlayerHands()

	ret.state.PlayersReady = make(map[string]string)
	for _, player := range ret.state.PlayerOrder {
		ret.state.PlayersReady[player] = ""
	}

	fmt.Println("created bunga:", ret)
	return ret
}

func (b *bunga) getBackHands() map[string][]string {
	ret := map[string][]string{}
	for _, player := range b.state.PlayerOrder {
		hand := make([]string, len(b.state.PlayerHands[player]))
		for i := 0; i < len(b.state.PlayerHands[player]); i++ {
			hand[i] = Back
			if b.state.SaidBunga == player {
				hand[i] += BungHl
			}
		}
		ret[player] = hand
	}
	return ret
}

func (b *bunga) computeScores() {
	b.state.Scores = map[string]int{}
	// Calculate base scores
	for player, hand := range b.state.PlayerHands {
		score := 0
		for _, card := range hand {
			switch card[0] {
			case 'A':
				continue
			case '2', '3', '4', '5', '6', '7', '8', '9':
				num, _ := strconv.Atoi(string(card[0]))
				score += num
			case 'T', 'J', 'Q':
				score += 10
			case 'K':
				if card[1] == 'H' || card[1] == 'D' {
					score -= 1
				} else {
					score += 25
				}
			}
		}
		b.state.Scores[player] = score
	}
	// Calculate penalty if bunga called incorrectly
	if b.state.Scores[b.state.SaidBunga] >= 10 {
		b.state.Scores[b.state.SaidBunga] += 10
	}
	// Calculate penalty if bunga caller lost
	minScore := b.state.Scores[b.state.PlayerOrder[0]]
	minScorePlayer := b.state.PlayerOrder[0]
	for player, score := range b.state.Scores {
		if score < minScore {
			minScore = score
			minScorePlayer = player
		}
	}
	// Determine winner
	b.state.Winner = minScorePlayer
}

func (b *bunga) getUserStatesStartGame() map[string]bungaUserState {
	ret := map[string]bungaUserState{}
	// have a map of hands full of card backs to use for everyone else
	backHands := b.getBackHands()
	// for each player, construct their own userState, which is their view of the game
	for _, player := range b.state.PlayerOrder {
		// construct PlayerHands. Use the back hand if it's not the players own hand.
		playerHands := map[string][]string{}
		for _, playerHand := range b.state.PlayerOrder {
			// player gets to see their own hand's first two cards
			// else they see a back hand
			if player == playerHand && b.state.PlayersReady[player] == "" {
				ownHand := []string{}
				for i, card := range b.state.PlayerHands[player] {
					if i < 2 {
						ownHand = append(ownHand, card+PrimHl)
					} else {
						ownHand = append(ownHand, Back)
					}
				}
				playerHands[player] = ownHand
			} else {
				playerHands[playerHand] = backHands[playerHand]
			}
		}
		ret[player] = bungaUserState{
			DrawPile:     Back,
			DiscardPile:  Blank,
			Turn:         "",
			PlayersReady: b.state.PlayersReady,
			PlayerHands:  playerHands,
			PlayerOrder:  b.state.PlayerOrder,
		}
	}
	return ret
}

func (b *bunga) getUserStatesPlaying() map[string]bungaUserState {
	fmt.Println("Bunga getting user states, playingState", b.state.PlayingState)
	ret := map[string]bungaUserState{}
	// for each player, construct their view of the game
	// 'player' is the player that sees the state
	for _, player := range b.state.PlayerOrder {
		backHands := b.getBackHands()
		playerHands := map[string][]string{}
		drawPile := Back
		// discard pile is blank if empty, else it's the top card
		discardPile := b.discardTop()
		// for each other players' hand, determine visibility + highlights, only if it's the players' turn though
		// 'playerHand' is the player who's hand we want to set visibility for
		for _, playerHand := range b.state.PlayerOrder {
			playerHands[playerHand] = backHands[playerHand]
			if b.state.Turn == player && b.state.Turn == playerHand {
				// construct player hands based on playingState
				// if it's start turn it's all back, draw pile and discard pile highlighted if not empty
				switch b.state.PlayingState {
				case StartTurn:
					drawPile += PrimHl
					if discardPile != Blank {
						discardPile += PrimHl
					}
				case DiscardSwapChoice:
					// if it's discardSwapChoice, hand gets highlighted
					for i, card := range playerHands[player] {
						playerHands[player][i] = card + PrimHl
					}
				case DrawChoice:
					// if it's drawChoice draw pile is up and hand and discard gets highlighted
					for i, card := range playerHands[player] {
						playerHands[player][i] = card + PrimHl
					}
					fmt.Println("Bunga draw pile:", strings.Join(b.state.DrawPile, ","))
					fmt.Println("Bunga draw top", b.drawTop())
					drawPile = b.drawTop()
					discardPile += PrimHl
				case LookOwnChoice:
					for i := range playerHands[player] {
						playerHands[player][i] += PrimHl
					}
				case LookingOwn:
					for i, card := range b.state.PlayerHands[player] {
						if card[:2] == b.state.CardSelected[:2] {
							playerHands[player][i] = card + PrimHl
						} else {
							playerHands[player][i] += PrimHl
						}
					}
				case SwapOtherOwnChoice:
					for i := range playerHands[player] {
						playerHands[player][i] += PrimHl
					}
				case LookSwapOwnChoice:
					for i := range playerHands[player] {
						playerHands[player][i] += PrimHl
					}
				}
			}
			if b.state.Turn == player && b.state.Turn != playerHand {
				// here we worry about setting visibility for other players' hands
				// e.g. for 9, 10, J, and Q
				switch b.state.PlayingState {
				case LookOtherChoice:
					for i := range playerHands[playerHand] {
						playerHands[playerHand][i] += PrimHl
					}
				case LookingOther:
					fmt.Println("getting vis for lookingOther, hand:", b.state.PlayerHands[playerHand])
					fmt.Println("getting vis card selected:", b.state.CardSelected)
					for i, card := range b.state.PlayerHands[playerHand] {
						if card[:2] == b.state.CardSelected[:2] {
							playerHands[playerHand][i] = card + PrimHl
						} else {
							playerHands[playerHand][i] += PrimHl
						}
					}
				case SwapOtherChoice:
					for i := range playerHands[playerHand] {
						playerHands[playerHand][i] += PrimHl
					}
				case SwapOtherOwnChoice:
					for i, card := range b.state.PlayerHands[playerHand] {
						if card[:2] == b.state.CardSelected[:2] {
							playerHands[playerHand][i] += SecoHl
						}
					}
				case LookSwapChoice:
					for i := range playerHands[playerHand] {
						playerHands[playerHand][i] += PrimHl
					}
				case LookSwapOwnChoice:
					for i, card := range b.state.PlayerHands[playerHand] {
						if card[:2] == b.state.CardSelected[:2] {
							playerHands[playerHand][i] = card + PrimHl
						} else {
							playerHands[playerHand][i] += PrimHl
						}
					}
				}
			}
		}
		ret[player] = bungaUserState{
			DrawPile:     drawPile,
			DiscardPile:  discardPile,
			LatestAction: b.state.LatestAction,
			Turn:         b.state.Turn,
			PlayerHands:  playerHands,
			PlayerOrder:  b.state.PlayerOrder,
			SaidBunga:    b.state.SaidBunga,
			PlayingState: b.state.PlayingState,
		}
	}
	b.state.LatestAction = []bungaAction{}
	return ret
}

func (b *bunga) getUserStatesEndGame() map[string]bungaUserState {
	b.computeScores()
	ret := map[string]bungaUserState{}
	ret[Final] = bungaUserState{
		DrawPile:    Back,
		DiscardPile: b.discardTop(),
		Turn:        Final,
		PlayerHands: b.state.PlayerHands,
		PlayerOrder: b.state.PlayerOrder,
		Scores:      b.state.Scores,
		Winner:      b.state.Winner,
	}
	return ret
}

// compute visibility based on game state
// also compute highlight status
func (b *bunga) broadcastState() {
	// call a getUserStates function depending on game state
	// for each player in the lobby, send them their state
	var userStates map[string]bungaUserState
	switch b.state.GameState {
	case StartGame:
		userStates = b.getUserStatesStartGame()
	case Playing:
		userStates = b.getUserStatesPlaying()
	case EndGame:
		userStates = b.getUserStatesEndGame()
	}

	if b.state.GameState != EndGame {
		for _, player := range b.state.PlayerOrder {
			b.out <- gameMsg{player, userStates[player]}
		}
	} else {
		b.out <- gameMsg{Final, userStates[Final]}
	}
}

// state machine ish functions to update game state

func (b *bunga) moveStartgameState(msg userMsg) {
	// Check if player is readying
	player := msg.Args["player"]
	idx, _ := strconv.Atoi(msg.Args["index"])
	if player == msg.Args["owner"] && idx < 2 {
		b.state.PlayersReady[player] = Ready
	}
	// If all players are ready, go to playing state
	allReady := true
	for _, readyState := range b.state.PlayersReady {
		readyBool := false
		if readyState == Ready {
			readyBool = true
		}
		allReady = allReady && readyBool
	}
	// Prep for starting game
	if allReady {
		fmt.Println("Bunga all players ready, starting game!")
		b.state.GameState = Playing
		b.state.PlayingState = StartTurn
		b.state.Turn = b.state.PlayerOrder[0]
	}
}

func (b *bunga) movePlayingState(msg userMsg) {
	// parse message for convenience
	player := msg.Args[Player]
	var idx int
	var idxStr string
	var card string
	var choseOwn bool
	var choseOther bool
	if msg.Cmd == Card {
		choseOwn = player == msg.Args[Owner]
		choseOther = player != msg.Args[Owner] && b.state.SaidBunga != msg.Args[Owner]
		idxStr = msg.Args[Index]
		idx, _ = strconv.Atoi(msg.Args[Index])
		card = b.state.PlayerHands[msg.Args[Owner]][idx]
	}
	// handle tagging logic
	notAllowedTagStates := map[string]struct{}{
		DiscardSwapChoice:  {},
		DrawChoice:         {},
		LookOwnChoice:      {},
		LookingOwn:         {},
		SwapOtherOwnChoice: {},
		LookSwapOwnChoice:  {},
	}
	// your turn     not allowed state   allowed tag
	//    false           false             true
	//    false           true              true
	//    true            false             true
	//    true            true              false
	_, disallowed := notAllowedTagStates[b.state.PlayingState]
	tagAllowed := b.state.Turn != player || !disallowed
	if choseOwn && tagAllowed {
		fmt.Println(player, "sent cmd 'Card' and tagAllowed:", tagAllowed)
		if len(b.state.PlayerHands[player]) > 1 && player != b.state.SaidBunga {
			canTag := true
			// check if it's the right card
			if card[0] != b.discardTop()[0] {
				canTag = false
			}
			// check if the top of the discard pile is the last tagged card
			// if so, we can't tag
			if b.state.LatestTag != "" && b.state.LatestTag[0] == card[0] {
				canTag = false
			}
			b.state.LatestAction = []bungaAction{
				{Start: player, StartIdx: idxStr, End: Discard},
			}
			// handle incorrect tag
			if !canTag {
				// add another card to hand
				b.state.PlayerHands[player] = append(b.state.PlayerHands[player], b.drawCard())
				// set latest actions
				b.state.LatestAction = append(b.state.LatestAction,
					bungaAction{
						Start: Discard, End: Discard,
						Card: Wrong,
					},
					bungaAction{
						Start: Draw, End: player,
						EndIdx: strconv.Itoa(len(b.state.PlayerHands[player]) - 1),
					},
				)
			} else {
				// move card to discard pile
				hand := b.state.PlayerHands[player]
				b.state.PlayerHands[player] = append(hand[:idx], hand[idx+1:]...)
				b.state.DiscardPile = append(b.state.DiscardPile, card)
				// set b.state.LatestTag to the tagged card
				b.state.LatestTag = card
			}
		}
	}
	// check if it's the right player
	if b.state.Turn != player {
		return
	}
	// based on the playingState, advance the state machine based on the given move
	switch b.state.PlayingState {
	case StartTurn:
		if msg.Cmd == Draw {
			b.state.PlayingState = DrawChoice
		} else if msg.Cmd == Discard && len(b.state.DiscardPile) > 0 {
			b.state.PlayingState = DiscardSwapChoice
		} else if msg.Cmd == Bunga && b.state.SaidBunga == "" {
			b.state.SaidBunga = player
			b.advanceTurn()
			b.state.PlayingState = StartTurn
		}
	case DrawChoice:
		if msg.Cmd == Discard || choseOwn {
			cardType := b.drawTop()[0]
			notSpecialTypes := map[byte]struct{}{
				'A': {}, '2': {}, '3': {}, '4': {}, '5': {}, '6': {}, 'K': {},
			}
			if len(b.state.PlayerOrder) < 2 && b.state.SaidBunga != "" {
				for _, t := range []byte{'9', 'T', 'J', 'Q'} {
					notSpecialTypes[t] = struct{}{}
				}
			}
			if msg.Cmd == Discard {
				b.state.DiscardPile = append(b.state.DiscardPile, b.drawCard())
				b.state.LatestAction = []bungaAction{
					{Start: Draw, End: Discard},
				}
				if _, notSpecial := notSpecialTypes[cardType]; notSpecial {
					b.advanceTurn()
					b.state.PlayingState = StartTurn
				} else {
					switch cardType {
					case '7', '8':
						b.state.PlayingState = LookOwnChoice
					case '9', 'T':
						if len(b.state.PlayerOrder) > 1 {
							b.state.PlayingState = LookOtherChoice
						} else {
							b.state.PlayingState = LookOwnChoice
						}
					case 'J':
						b.state.PlayingState = SwapOtherChoice
					case 'Q':
						b.state.PlayingState = LookSwapChoice
					}
				}
			} else {
				// top of draw pile -> clicked card
				// clicked card -> discard pile
				b.state.DiscardPile = append(b.state.DiscardPile, card)
				b.state.PlayerHands[player][idx] = b.drawCard()
				b.state.LatestAction = []bungaAction{
					{Start: Draw, End: player, EndIdx: idxStr},
					{Start: player, StartIdx: idxStr, End: Discard},
				}
				b.advanceTurn()
				b.state.PlayingState = StartTurn
			}
		}
	case DiscardSwapChoice:
		if choseOwn {
			// Swap the discard top and the clicked card
			b.state.PlayerHands[player][idx] = b.state.DiscardPile[len(b.state.DiscardPile)-1]
			b.state.DiscardPile[len(b.state.DiscardPile)-1] = card
			b.state.LatestAction = []bungaAction{
				{Start: Discard, End: player, EndIdx: idxStr},
				{Start: player, StartIdx: idxStr, End: Discard},
			}
			b.advanceTurn()
			b.state.PlayingState = StartTurn
		}
	case LookOwnChoice:
		if choseOwn {
			b.state.CardSelected = card
			b.state.PlayingState = LookingOwn
		}
	case LookingOwn:
		if choseOwn {
			b.state.CardSelected = ""
			b.advanceTurn()
			b.state.PlayingState = StartTurn
		}
	case LookOtherChoice:
		if choseOther {
			b.state.CardSelected = card
			fmt.Println("Looking other, card selected:", card)
			b.state.PlayingState = LookingOther
		}
	case LookingOther:
		if choseOther {
			b.state.CardSelected = ""
			b.advanceTurn()
			b.state.PlayingState = StartTurn
		}
	case SwapOtherChoice:
		if choseOther {
			b.state.CardSelected = card
			b.state.PlayingState = SwapOtherOwnChoice
		}
	case SwapOtherOwnChoice:
		if choseOwn {
			for otherPlayer, hand := range b.state.PlayerHands {
				for i, otherCard := range hand {
					if otherCard == b.state.CardSelected {
						b.state.PlayerHands[otherPlayer][i] = card
						b.state.PlayerHands[player][idx] = otherCard
						iStr := strconv.Itoa(i)
						b.state.LatestAction = []bungaAction{
							{Start: player, StartIdx: idxStr, End: otherPlayer, EndIdx: iStr},
							{Start: otherPlayer, StartIdx: iStr, End: player, EndIdx: idxStr},
						}
						b.state.CardSelected = ""
						b.advanceTurn()
						b.state.PlayingState = StartTurn
					}
				}
			}
		}
	case LookSwapChoice:
		if choseOther {
			b.state.CardSelected = card
			b.state.PlayingState = LookSwapOwnChoice
		}
	case LookSwapOwnChoice:
		if choseOther {
			b.state.CardSelected = ""
			b.advanceTurn()
			b.state.PlayingState = StartTurn
		}
		if choseOwn {
			for otherPlayer, hand := range b.state.PlayerHands {
				for i, otherCard := range hand {
					if otherCard == b.state.CardSelected {
						b.state.PlayerHands[otherPlayer][i] = card
						b.state.PlayerHands[player][idx] = otherCard
						iStr := strconv.Itoa(i)
						b.state.LatestAction = []bungaAction{
							{Start: player, StartIdx: idxStr, End: otherPlayer, EndIdx: iStr},
							{Start: otherPlayer, StartIdx: iStr, End: player, EndIdx: idxStr},
						}
						b.state.CardSelected = ""
						b.advanceTurn()
						b.state.PlayingState = StartTurn
					}
				}
			}
		}
	}
}

func (b *bunga) runGame() {
	fmt.Println("Bunga starting!")
	go b.broadcastState()
	for msg := range b.in {
		fmt.Println("Bunga got message, processing...")

		fmt.Println(msg.Cmd, msg.Args)
		if b.state.GameState == StartGame {
			b.moveStartgameState(msg)
		} else if b.state.GameState == Playing {
			b.movePlayingState(msg)
		}

		b.broadcastState()
		if b.state.GameState == EndGame {
			fmt.Println("Bunga done")
			return
		}
	}
}
