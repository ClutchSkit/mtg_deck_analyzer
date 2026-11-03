package analysis

import (
	"math"
	"testing"
)

func TestCombinations(t *testing.T) {
	combination := combinations(5, 2)
	if combination != 10 {
		t.Fatalf("Expected combinations(5, 2) to be 10, got %f", combination)
	}
	combination = combinations(10, 8)
	if combination != 45 {
		t.Fatalf("Expected combinations(10, 8) to be 45, got %f", combination)
	}
}

func TestHypergeometricProbability(t *testing.T) {
	probability, err := HypergeometricProbability(99, 36, 2)
	if err != nil {
		t.Fatalf("Expected HypergeometricProbability(99, 36, 2) to succeed, got error: %v", err)
	}
	expected := 0.861
	if math.Abs(probability-expected) > 0.001 {
		t.Fatalf("Expected HypergeometricProbability(99, 36, 2) to be %f, got %f", expected, probability)
	}

	probability, err = HypergeometricProbability(99, 10, 1)
	if err != nil {
		t.Fatalf("Expected HypergeometricProbability(99, 10, 1) to succeed, got error: %v", err)
	}
	expected = 0.587
	if math.Abs(probability-expected) > 0.001 {
		t.Fatalf("Expected HypergeometricProbability(99, 10, 1) to be %f, got %f", expected, probability)
	}
}
