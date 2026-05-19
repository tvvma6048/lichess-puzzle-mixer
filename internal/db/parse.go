package db

import "strings"

// solverSide returns the side the human solver plays (opposite FEN side to move).
func solverSide(fen string) string {
	parts := strings.Fields(fen)
	if len(parts) < 2 {
		return "w"
	}
	switch parts[1] {
	case "b":
		return "w"
	case "w":
		return "b"
	default:
		return "w"
	}
}

func solutionLength(moves string) int {
	n := len(strings.Fields(moves))
	if n <= 1 {
		return 0
	}
	return n - 1
}

func splitThemes(raw string) []string {
	if raw == "" {
		return nil
	}
	return strings.Fields(raw)
}

func parseOpeningTags(raw string) (family, variation *string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Fields(raw)
	f := parts[0]
	family = &f
	variation = &raw
	return family, variation
}

func splitMoves(raw string) []string {
	return strings.Fields(raw)
}
