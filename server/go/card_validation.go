package main

import "errors"

func validateCards(cards []card) error {
	if len(cards) != 2 {
		return errors.New("cards must contain exactly 2 cards")
	}
	for _, card := range cards {
		if !isValidRank(card.Rank) {
			return errors.New("rank must be one of 2-9, T, J, Q, K, A")
		}
		if !isValidSuit(card.Suit) {
			return errors.New("suit must be one of s, h, d, c")
		}
	}
	if cards[0].Rank == cards[1].Rank && cards[0].Suit == cards[1].Suit {
		return errors.New("cards must not contain the same card twice")
	}
	return nil
}

func isValidRank(rank string) bool {
	return len(rank) == 1 && contains("23456789TJQKA", rank)
}

func isValidSuit(suit string) bool {
	return len(suit) == 1 && contains("shdc", suit)
}

func contains(values, value string) bool {
	for _, candidate := range values {
		if string(candidate) == value {
			return true
		}
	}
	return false
}
