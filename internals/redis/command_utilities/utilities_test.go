package commandutilities

import "testing"

func TestIncrementConnCounter(t *testing.T) {
	t.Run("increment conn counter", func(t *testing.T) {
		InitializeCounter()
		IncrementConnCounter()
		if GetConnCounter() != 1 {
			t.Errorf("Expected 1, got %d", GetConnCounter())
		}
	})

	t.Run("increment conn counter multiple times", func(t *testing.T) {
		InitializeCounter()
		IncrementConnCounter()
		IncrementConnCounter()
		if GetConnCounter() != 2 {
			t.Errorf("Expected 2, got %d", GetConnCounter())
		}
	})
}

func TestDecrementConnCounter(t *testing.T) {
	t.Run("decrement conn counter", func(t *testing.T) {
		InitializeCounterWithValue(1)
		DecrementConnCounter()
		if GetConnCounter() != 0 {
			t.Errorf("Expected 0, got %d", GetConnCounter())
		}
	})

	t.Run("decrement conn counter multiple times", func(t *testing.T) {
		InitializeCounterWithValue(2)
		DecrementConnCounter()
		DecrementConnCounter()
		if GetConnCounter() != 0 {
			t.Errorf("Expected 0, got %d", GetConnCounter())
		}
	})

	t.Run("decrement conn counter when counter is 0", func(t *testing.T) {
		InitializeCounterWithValue(0)
		DecrementConnCounter()
		if GetConnCounter() != 0 {
			t.Errorf("Expected 0, got %d", GetConnCounter())
		}
	})
}
