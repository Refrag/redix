package commandutilities_test

import (
	"testing"

	cmdUtils "github.com/Refrag/redix/internals/redis/command_utilities"
)

func TestIncrementConnCounter(t *testing.T) {
	t.Run("increment conn counter", func(t *testing.T) {
		cmdUtils.InitializeCounter()
		cmdUtils.IncrementConnCounter()
		if cmdUtils.GetConnCounter() != 1 {
			t.Errorf("Expected 1, got %d", cmdUtils.GetConnCounter())
		}
	})

	t.Run("increment conn counter multiple times", func(t *testing.T) {
		cmdUtils.InitializeCounter()
		cmdUtils.IncrementConnCounter()
		cmdUtils.IncrementConnCounter()
		if cmdUtils.GetConnCounter() != 2 {
			t.Errorf("Expected 2, got %d", cmdUtils.GetConnCounter())
		}
	})
}

func TestDecrementConnCounter(t *testing.T) {
	t.Run("decrement conn counter", func(t *testing.T) {
		cmdUtils.InitializeCounterWithValue(1)
		cmdUtils.DecrementConnCounter()
		if cmdUtils.GetConnCounter() != 0 {
			t.Errorf("Expected 0, got %d", cmdUtils.GetConnCounter())
		}
	})

	t.Run("decrement conn counter multiple times", func(t *testing.T) {
		cmdUtils.InitializeCounterWithValue(2)
		cmdUtils.DecrementConnCounter()
		cmdUtils.DecrementConnCounter()
		if cmdUtils.GetConnCounter() != 0 {
			t.Errorf("Expected 0, got %d", cmdUtils.GetConnCounter())
		}
	})

	t.Run("decrement conn counter when counter is 0", func(t *testing.T) {
		cmdUtils.InitializeCounterWithValue(0)
		cmdUtils.DecrementConnCounter()
		if cmdUtils.GetConnCounter() != 0 {
			t.Errorf("Expected 0, got %d", cmdUtils.GetConnCounter())
		}
	})
}
