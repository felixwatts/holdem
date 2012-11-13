package holdem

import "math/rand"

type Hand uint64
type Card uint8
type Suit uint8
type Face uint8

const (
	NUM_CARDS = 52

	S Suit = 0
	H Suit = 1
	D Suit = 2
	C Suit = 3

	C2  Face = 0
	C3  Face = 1
	C4  Face = 2
	C5  Face = 3
	C6  Face = 4
	C7  Face = 5
	C8  Face = 6
	C9  Face = 7
	C10 Face = 8
	J   Face = 9
	Q   Face = 10
	K   Face = 11
	A   Face = 12
)

// Compare can be used to determine which of two hands is better
// taking into account ways in which each can develop as further
// cards are drawn. Given two hands h1, h2 and the total number of cards
// that will be in each hand at end of play, it returns the chance
// that h1 will beat h2 at end of play.
// For example, in a game of texas holdem, each player effectively has
// 7 cards after the river. Their hand is the best hand they can make
// with those 7 cards. After the flop, you know 5 or your 7 cards and
// 3 of the 7 cards of your opponent. You can use Compare to determine
// the chance that your hand will beat your opponent's after the river
// using: Compare(myHand|table, table, 7)
func Compare(h1 Hand, h2 Hand, lookahead uint8) float64 {
	l1 := h1.NumCards()
	l2 := h2.NumCards()

	var s uint8
	switch {
	case l1 > l2:
		s = lookahead - l2
	case l1 < l2:
		s = lookahead - l1
	default:
		s = lookahead - l1
	}

	return _compare(h1, l1, h2, l2, Card(0), s)
}

func _compare(h1 Hand, l1 uint8, h2 Hand, l2 uint8, firstCard Card, lookahead uint8) float64 {
	if lookahead == 0 {
		s1 := h1.Score()
		s2 := h2.Score()
		switch {
		case s1 > s2:
			return 1.0
		case s1 < s2:
			return 0.0
		default:
			return 0.5
		}
	} else {

		ch := 0.0
		nc := 0
		if l1 < l2 {
			// add a card to h1
			for c := firstCard; c < NUM_CARDS; c++ {
				if !h1.HasCard(c) && !h2.HasCard(c) {
					ch += _compare(h1.AddCard(c), l1+1, h2, l2, c+1, lookahead-1)
					nc++
				}
			}
		} else if l2 < l1 {
			// add a card to h2
			for c := firstCard; c < NUM_CARDS; c++ {
				if !h1.HasCard(c) && !h2.HasCard(c) {
					ch += _compare(h1, l1, h2.AddCard(c), l2+1, c+1, lookahead-1)
					nc++
				}

			}
		} else {
			// add a card to both
			for c := firstCard; c < NUM_CARDS; c++ {
				if !h1.HasCard(c) && !h2.HasCard(c) {
					ch += _compare(h1.AddCard(c), l1+1, h2.AddCard(c), l2+1, c+1, lookahead-1)
					nc++
				}
			}
		}

		if nc == 0 {
			return 0.5
		}

		ch /= float64(nc)

		return ch
	}

	return 0
}

// Describe provides an english description of a hand, such as 'two pairs'
func (h Hand) Describe() string {

	fc := FaceCounts(h)

	switch {
	case straightFlushScore(h) == 13:
		return "royal flush"
	case straightFlushScore(h) != 0:
		return "straight flush"
	case fullHouseScore(fc) != 0:
		return "full house"
	case flushScore(h) != 0:
		return "flush"
	case threeKindScore(fc) != 0:
		return "three of a kind"
	case twoPairScore(fc) != 0:
		return "two pairs"
	case pairScore(fc) != 0:
		return "pair"
	}

	return "high Card"
}

// Score returns a number representing the rank of a hand among
// all possible hands. If h1.Score() > h2.Score() then h1 beats
// h2. Currently tie breaking for equivalent hands is not handled,
// for example, 🂡🂱🃂🃒🃓 scores the same as 🂡🂱🃂🃒🃔
func (h Hand) Score() uint32 {
	s := straightFlushScore(h)
	if s == 0 {
		fc := FaceCounts(h)
		s = fourKindScore(fc)
		if s == 0 {
			s = fullHouseScore(fc)
			if s == 0 {
				s = flushScore(h)
				if s == 0 {
					s = straightScore(h)
					if s == 0 {
						s = threeKindScore(fc)
						if s == 0 {
							s = twoPairScore(fc)
							if s == 0 {
								s = pairScore(fc)
								if s == 0 {
									return uint32(highCardScore(fc))
								} else {
									return uint32(s) + 13
								}
							} else {
								return uint32(s) + (2 * 13)
							}
						} else {
							return uint32(s) + (3 * 13)
						}
					} else {
						return uint32(s) + (4 * 13)
					}
				} else {
					return uint32(s) + (5 * 13)
				}
			} else {
				return uint32(s) + (6 * 13)
			}
		} else {
			return uint32(s) + (7 * 13)
		}
	} else {
		return uint32(s) + (8 * 13)
	}

	return 0
}

func straightFlushScore(h Hand) Face {
	for s := Suit(0); s < 4; s++ {
		runLength := 0
		for f := Face(0); f < 13; f++ {
			if h.HasCard(ToCard(f, s)) {
				runLength++
				if runLength == 5 {
					return f + 1
				}
			} else {
				runLength = 0
			}
		}
	}
	return 0
}

func flushScore(h Hand) Face {
	SuitCount := make([]Card, 4)
	SuitMax := make([]Face, 4)
	for c := Card(0); c < NUM_CARDS; c++ {
		if h.HasCard(c) {
			s := c.Suit()
			f := c.Face()
			SuitCount[s]++
			if SuitMax[s] < f+1 {
				SuitMax[s] = f + 1
			}
			if SuitCount[s] == 5 {
				return SuitMax[s]
			}
		}
	}
	return 0
}

func straightScore(h Hand) Face {
	runLength := 0
	max := Face(0)
	for f := Face(12); f <= 12; f-- {
		var s Suit
		found := false

		for s = Suit(0); !found && s < 4; s++ {
			if h.HasCard(ToCard(f, s)) {
				if runLength == 0 {
					max = f + 1
				}
				runLength++
				if runLength == 5 {
					return max
				}
				found = true
				break
			}
		}

		if !found {
			runLength = 0
			max = 0
		}
	}
	return 0
}

func fourKindScore(fc []uint8) Face {
	for f := Face(12); f <= 12; f-- {
		if fc[f] == 4 {
			return f + 1
		}
	}
	return 0
}

func fullHouseScore(fc []uint8) Face {
	max := Face(0)
	hasTwo := false
	for f := Face(0); f < 13; f++ {
		if fc[f] == 2 {
			hasTwo = true
		}
		if fc[f] == 3 && f > max {
			max = f + 1
		}
	}
	if hasTwo {
		return max
	}
	return 0
}

func threeKindScore(fc []uint8) Face {
	max := Face(0)
	for f := Face(0); f < 13; f++ {
		if fc[f] == 3 && f > max {
			max = f + 1
		}
	}
	return max
}

func twoPairScore(fc []uint8) Face {
	max := Face(0)
	pairCount := 0
	for f := Face(0); f < 13; f++ {
		if fc[f] == 2 {
			pairCount++
			if f+1 > max {
				max = f + 1
			}
		}
	}
	if pairCount >= 2 {
		return max
	}
	return 0
}

func pairScore(fc []uint8) Face {
	max := Face(0)
	for f := Face(0); f < 13; f++ {
		if fc[f] == 2 && f+1 > max {
			max = f + 1
		}
	}
	return max
}

func highCardScore(fc []uint8) Face {
	for f := Face(12); f <= 12; f-- {
		if fc[f] == 1 {
			return f + 1
		}
	}
	return 0
}

// FaceCounts counts the occurrences of each face within a hand.
func FaceCounts(h Hand) []uint8 {
	result := make([]uint8, 13)
	for f := Face(0); f < 13; f++ {
		for s := Suit(0); s < 4; s++ {
			if h.HasCard(ToCard(f, s)) {
				result[f]++
			}
		}
	}
	return result
}

// HasCard returns true of the hand contains the specified Card.
func (h Hand) HasCard(c Card) bool {
	return (h & (1 << c)) != 0
}

// Suit returns the Suit of the gioven Card
func (c Card) Suit() Suit {
	return Suit(c / 13)
}

// Face returns the Face of the given Card
func (c Card) Face() Face {
	return Face(c % 13)
}

// ToCard crates a card from Suit and Face.
func ToCard(f Face, s Suit) Card {
	return Card(uint8(s)*13 + uint8(f))
}

// AddCard returns the result of adding the given card to the given
// hand. The original Hand is not effected. A Hand cannot contain
// the same Card multiple times.
func (h Hand) AddCard(c Card) Hand {
	return h | (1 << c)
}

// CreateHand returns a hand containing the given cards.
func CreateHand(cards ...Card) Hand {
	h := Hand(0)
	for _, v := range cards {
		h = h.AddCard(v)
	}
	return h
}

// NumCards counts the cards in the given Hand
func (h Hand) NumCards() uint8 {
	result := uint8(0)
	for c := Card(0); c < NUM_CARDS; c++ {
		if h.HasCard(c) {
			result++
		}
	}
	return result
}

// Combine returns the result of combining the two
// given hands. The original hands are not effected.
func (h1 Hand) Combine(h2 Hand) Hand {
	return h1 | h2
}

// RandomHand returns a hand of the specified size
// comprising psuedorandomly selected cards.
func RandomHand(size uint8) Hand {
	h := Hand(0)
	for h.NumCards() < size {
		c := Card(rand.Int31n(NUM_CARDS))
		h = h.AddCard(c)
	}
	return h
}

func (h Hand) String() string {
	result := ""

	for f := Face(0); f < 13; f++ {
		for s := Suit(0); s < 4; s++ {
			c := ToCard(f, s)
			if h.HasCard(c) {
				result += c.String()
			}
		}
	}

	return result
}

func (c Card) String() string {

	fs := c.Face()
	if fs == A {
		fs = 0
	} else if fs < Q {
		fs++
	} else {
		fs += 2
	}

	ss := c.Suit()

	return string('🂡'+(uint(fs)*0x1)+(uint(ss)*0x10)) + " "
}
