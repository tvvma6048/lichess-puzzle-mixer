package db

import "testing"

func TestSolverSide(t *testing.T) {
	fen := "r6k/pp2r2p/4Rp1Q/3p4/8/1N1P2R1/PqP2bPP/7K b - - 0 24"
	if got := solverSide(fen); got != "w" {
		t.Fatalf("solverSide = %q, want w", got)
	}
}

func TestSolutionLength(t *testing.T) {
	moves := "f2g3 e6e7 b2b1 b3c1 b1c1 h6c1"
	if got := solutionLength(moves); got != 5 {
		t.Fatalf("solutionLength = %d, want 5", got)
	}
}
