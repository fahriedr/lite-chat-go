package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	t.Run("Generate random string of specified length", func(t *testing.T) {
		lengths := []int{5, 10, 20, 50}

		for _, length := range lengths {
			result := RandomString(length)
			assert.Equal(t, length, len(result))
			assert.NotEmpty(t, result)
		}
	})

	t.Run("Random strings are different", func(t *testing.T) {
		str1 := RandomString(10)
		str2 := RandomString(10)
		
		// They should be different (very high probability)
		assert.NotEqual(t, str1, str2)
	})

	t.Run("Zero length string", func(t *testing.T) {
		result := RandomString(0)
		assert.Equal(t, 0, len(result))
		assert.Equal(t, "", result)
	})

	t.Run("Single character string", func(t *testing.T) {
		result := RandomString(1)
		assert.Equal(t, 1, len(result))
		assert.NotEmpty(t, result)
	})
}

func TestEmailToUsername(t *testing.T) {
	t.Run("Valid email to username", func(t *testing.T) {
		email := "test@example.com"
		result := EmailToUsername(email)
		
		assert.Contains(t, result, "test")
		assert.True(t, len(result) > 4) // Should have random digits added
		assert.NotEqual(t, "test", result) // Should be different from just the local part
	})

	t.Run("Complex email to username", func(t *testing.T) {
		email := "john.doe+label@gmail.com"
		result := EmailToUsername(email)
		
		assert.Contains(t, result, "john.doe+label")
		assert.True(t, len(result) > len("john.doe+label"))
	})

	t.Run("Multiple calls generate different usernames", func(t *testing.T) {
		email := "same@example.com"
		result1 := EmailToUsername(email)
		result2 := EmailToUsername(email)
		
		// Should generate different numbers
		assert.NotEqual(t, result1, result2)
		assert.Contains(t, result1, "same")
		assert.Contains(t, result2, "same")
	})

	t.Run("Email with numbers", func(t *testing.T) {
		email := "user123@example.com"
		result := EmailToUsername(email)
		
		assert.Contains(t, result, "user123")
		assert.True(t, len(result) > len("user123"))
	})
}

func TestMapToJSON(t *testing.T) {
	t.Run("Simple map to JSON", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "John",
			"age":  30,
		}

		result, err := MapToJSON(data)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "name")
		assert.Contains(t, result, "John")
		assert.Contains(t, result, "age")
		assert.Contains(t, result, "30")
	})

	t.Run("Empty map to JSON", func(t *testing.T) {
		data := map[string]interface{}{}

		result, err := MapToJSON(data)
		
		assert.NoError(t, err)
		assert.Equal(t, "{}", result)
	})

	t.Run("Nested map to JSON", func(t *testing.T) {
		data := map[string]interface{}{
			"user": map[string]interface{}{
				"name":  "John",
				"email": "john@example.com",
			},
			"active": true,
		}

		result, err := MapToJSON(data)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "user")
		assert.Contains(t, result, "name")
		assert.Contains(t, result, "John")
		assert.Contains(t, result, "active")
		assert.Contains(t, result, "true")
	})

	t.Run("Map with various data types", func(t *testing.T) {
		data := map[string]interface{}{
			"string":  "text",
			"number":  42,
			"float":   3.14,
			"boolean": true,
			"null":    nil,
			"array":   []string{"a", "b", "c"},
		}

		result, err := MapToJSON(data)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "string")
		assert.Contains(t, result, "number")
		assert.Contains(t, result, "float")
		assert.Contains(t, result, "boolean")
		assert.Contains(t, result, "array")
	})

	t.Run("Map with invalid JSON data", func(t *testing.T) {
		// Create a map with a channel (which can't be JSON marshaled)
		data := map[string]interface{}{
			"invalid": make(chan int),
		}

		result, err := MapToJSON(data)
		
		assert.Error(t, err)
		assert.Empty(t, result)
	})
}

// Test utility functions for validation
func TestValidate(t *testing.T) {
	t.Run("Validate variable should be initialized", func(t *testing.T) {
		assert.NotNil(t, Validate)
	})
}

// Benchmark tests
func BenchmarkRandomString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RandomString(10)
	}
}

func BenchmarkEmailToUsername(b *testing.B) {
	email := "test@example.com"
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		EmailToUsername(email)
	}
}

func BenchmarkMapToJSON(b *testing.B) {
	data := map[string]interface{}{
		"name": "John",
		"age":  30,
		"active": true,
	}
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		MapToJSON(data)
	}
}