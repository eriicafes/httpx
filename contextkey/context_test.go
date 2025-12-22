package contextkey_test

import (
	"context"
	"slices"
	"testing"

	"github.com/eriicafes/httpx/contextkey"
)

func TestKey_Set_Get(t *testing.T) {
	key := contextkey.New[string]("test-key")
	ctx := context.Background()

	// Test getting from empty context
	value, ok := key.Get(ctx)
	if ok {
		t.Error("Expected ok to be false for empty context")
	}
	if value != "" {
		t.Errorf("Expected zero value, got %v", value)
	}

	// Test setting and getting a value
	ctx = key.Set(ctx, "hello")
	value, ok = key.Get(ctx)
	if !ok {
		t.Error("Expected ok to be true after setting value")
	}
	if value != "hello" {
		t.Errorf("Expected 'hello', got %v", value)
	}

	// Test overwriting a value
	ctx = key.Set(ctx, "world")
	value, ok = key.Get(ctx)
	if !ok {
		t.Error("Expected ok to be true after overwriting value")
	}
	if value != "world" {
		t.Errorf("Expected 'world', got %v", value)
	}
}

func TestKey_Delete(t *testing.T) {
	key := contextkey.New[int]("test-key")
	ctx := context.Background()

	// Set a value
	ctx = key.Set(ctx, 42)
	value, ok := key.Get(ctx)
	if !ok || value != 42 {
		t.Fatal("Failed to set initial value")
	}

	// Delete the value
	ctx = key.Delete(ctx)
	value, ok = key.Get(ctx)
	if ok {
		t.Error("Expected ok to be false after delete")
	}
	if value != 0 {
		t.Errorf("Expected zero value after delete, got %v", value)
	}

	// Delete from empty context should be safe
	ctx = context.Background()
	ctx = key.Delete(ctx)
	value, ok = key.Get(ctx)
	if ok {
		t.Error("Expected ok to be false")
	}
}

func TestKey_Take(t *testing.T) {
	key := contextkey.New[string]("test-key")
	ctx := context.Background()

	// Take from empty context
	newCtx, value, ok := key.Take(ctx)
	if ok {
		t.Error("Expected ok to be false for empty context")
	}
	if value != "" {
		t.Errorf("Expected zero value, got %v", value)
	}
	if newCtx != ctx {
		t.Error("Expected same context when value not found")
	}

	// Set a value and take it
	ctx = key.Set(ctx, "take-me")
	newCtx, value, ok = key.Take(ctx)
	if !ok {
		t.Error("Expected ok to be true when value exists")
	}
	if value != "take-me" {
		t.Errorf("Expected 'take-me', got %v", value)
	}

	// Verify value is removed from new context
	value, ok = key.Get(newCtx)
	if ok {
		t.Error("Expected value to be removed after take")
	}
	if value != "" {
		t.Errorf("Expected zero value after take, got %v", value)
	}

	// Verify original context still has the value
	value, ok = key.Get(ctx)
	if !ok || value != "take-me" {
		t.Error("Original context should still have the value")
	}
}

func TestKey_Update(t *testing.T) {
	key := contextkey.New[int]("test-key")
	ctx := context.Background()

	// Update on empty context
	ctx = key.Update(ctx, func(prev int, hasPrev bool) int {
		if hasPrev {
			t.Error("Expected hasPrev to be false for empty context")
		}
		if prev != 0 {
			t.Errorf("Expected zero value, got %v", prev)
		}
		return 10
	})

	value, ok := key.Get(ctx)
	if !ok || value != 10 {
		t.Errorf("Expected value to be 10, got %v (ok=%v)", value, ok)
	}

	// Update existing value
	ctx = key.Update(ctx, func(prev int, hasPrev bool) int {
		if !hasPrev {
			t.Error("Expected hasPrev to be true")
		}
		if prev != 10 {
			t.Errorf("Expected prev to be 10, got %v", prev)
		}
		return prev + 5
	})

	value, ok = key.Get(ctx)
	if !ok || value != 15 {
		t.Errorf("Expected value to be 15, got %v (ok=%v)", value, ok)
	}
}

func TestKey_MultipleKeys(t *testing.T) {
	keyString := contextkey.New[string]("key-string")
	keyInt := contextkey.New[int]("key-int")
	keyBool := contextkey.New[bool]("key-bool")

	ctx := context.Background()
	ctx = keyString.Set(ctx, "hello")
	ctx = keyInt.Set(ctx, 42)
	ctx = keyBool.Set(ctx, true)

	// Verify all values are independent
	if val, ok := keyString.Get(ctx); !ok || val != "hello" {
		t.Errorf("Expected 'hello', got %v (ok=%v)", val, ok)
	}
	if val, ok := keyInt.Get(ctx); !ok || val != 42 {
		t.Errorf("Expected 42, got %v (ok=%v)", val, ok)
	}
	if val, ok := keyBool.Get(ctx); !ok || val != true {
		t.Errorf("Expected true, got %v (ok=%v)", val, ok)
	}

	// Delete one key shouldn't affect others
	ctx = keyInt.Delete(ctx)
	if val, ok := keyString.Get(ctx); !ok || val != "hello" {
		t.Error("Deleting keyInt affected keyString")
	}
	if val, ok := keyInt.Get(ctx); ok || val != 0 {
		t.Error("keyInt should be deleted")
	}
	if val, ok := keyBool.Get(ctx); !ok || val != true {
		t.Error("Deleting keyInt affected keyBool")
	}
}

func TestKey_TypeSafety(t *testing.T) {
	// Keys with the same name but different types should be independent
	keyString := contextkey.New[string]("same-name")
	keyInt := contextkey.New[int]("same-name")

	ctx := context.Background()
	ctx = keyString.Set(ctx, "text")
	ctx = keyInt.Set(ctx, 123)

	// Both should coexist independently
	strVal, ok := keyString.Get(ctx)
	if !ok || strVal != "text" {
		t.Errorf("Expected 'text', got %v (ok=%v)", strVal, ok)
	}

	intVal, ok := keyInt.Get(ctx)
	if !ok || intVal != 123 {
		t.Errorf("Expected 123, got %v (ok=%v)", intVal, ok)
	}
}

func TestKey_ComplexTypes(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	key := contextkey.New[User]("user-key")
	ctx := context.Background()

	user := User{Name: "Alice", Age: 30}
	ctx = key.Set(ctx, user)

	retrieved, ok := key.Get(ctx)
	if !ok {
		t.Fatal("Expected to retrieve user")
	}
	if retrieved != user {
		t.Errorf("Expected %+v, got %+v", user, retrieved)
	}
}

func TestKey_PointerTypes(t *testing.T) {
	key := contextkey.New[*string]("pointer-key")
	ctx := context.Background()

	// Test with nil pointer
	ctx = key.Set(ctx, nil)
	value, ok := key.Get(ctx)
	if !ok {
		t.Error("Expected ok to be true even with nil pointer")
	}
	if value != nil {
		t.Errorf("Expected nil, got %v", value)
	}

	// Test with actual pointer
	str := "hello"
	ctx = key.Set(ctx, &str)
	value, ok = key.Get(ctx)
	if !ok || value == nil || *value != "hello" {
		t.Errorf("Expected pointer to 'hello', got %v (ok=%v)", value, ok)
	}
}

func TestKey_ZeroValues(t *testing.T) {
	// Test that zero values can be distinguished from missing values
	tests := []struct {
		name      string
		setupFunc func(context.Context) context.Context
		checkFunc func(context.Context) bool
	}{
		{
			name: "zero int",
			setupFunc: func(ctx context.Context) context.Context {
				key := contextkey.New[int]("zero-int")
				return key.Set(ctx, 0)
			},
			checkFunc: func(ctx context.Context) bool {
				key := contextkey.New[int]("zero-int")
				val, ok := key.Get(ctx)
				return ok && val == 0
			},
		},
		{
			name: "empty string",
			setupFunc: func(ctx context.Context) context.Context {
				key := contextkey.New[string]("empty-string")
				return key.Set(ctx, "")
			},
			checkFunc: func(ctx context.Context) bool {
				key := contextkey.New[string]("empty-string")
				val, ok := key.Get(ctx)
				return ok && val == ""
			},
		},
		{
			name: "false bool",
			setupFunc: func(ctx context.Context) context.Context {
				key := contextkey.New[bool]("false-bool")
				return key.Set(ctx, false)
			},
			checkFunc: func(ctx context.Context) bool {
				key := contextkey.New[bool]("false-bool")
				val, ok := key.Get(ctx)
				return ok && val == false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = tt.setupFunc(ctx)
			if !tt.checkFunc(ctx) {
				t.Errorf("Failed to properly handle zero value for %s", tt.name)
			}
		})
	}
}

func TestKey_Update_Chaining(t *testing.T) {
	key := contextkey.New[[]string]("list-key")
	ctx := context.Background()

	// Initialize with empty slice
	ctx = key.Update(ctx, func(prev []string, hasPrev bool) []string {
		return []string{"first"}
	})

	// Append multiple times
	ctx = key.Update(ctx, func(prev []string, hasPrev bool) []string {
		return append(prev, "second")
	})

	ctx = key.Update(ctx, func(prev []string, hasPrev bool) []string {
		return append(prev, "third")
	})

	value, ok := key.Get(ctx)
	if !ok {
		t.Fatal("Expected value to exist")
	}
	if len(value) != 3 {
		t.Errorf("Expected slice length 3, got %d", len(value))
	}
	expected := []string{"first", "second", "third"}
	if !slices.Equal(value, expected) {
		t.Errorf("Expected %v, got %v", expected, value)
	}
}
