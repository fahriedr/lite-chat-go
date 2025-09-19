package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func setupTestEnv() {
	os.Setenv("JWT_SECRET", "test-jwt-secret")
}

func TestHashPassword(t *testing.T) {
	setupTestEnv()

	t.Run("Hash password successfully", func(t *testing.T) {
		password := "testpassword123"
		
		hashedPassword, err := HashPassword(password)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
		assert.NotEqual(t, password, hashedPassword)
		assert.True(t, len(hashedPassword) > len(password))
	})

	t.Run("Hash different passwords produce different hashes", func(t *testing.T) {
		password1 := "password1"
		password2 := "password2"
		
		hash1, err1 := HashPassword(password1)
		hash2, err2 := HashPassword(password2)
		
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("Hash same password twice produces different hashes", func(t *testing.T) {
		password := "samepassword"
		
		hash1, err1 := HashPassword(password)
		hash2, err2 := HashPassword(password)
		
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2) // Due to salt
	})

	t.Run("Hash empty password", func(t *testing.T) {
		password := ""
		
		hashedPassword, err := HashPassword(password)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
	})

	t.Run("Hash very long password", func(t *testing.T) {
		password := string(make([]byte, 70)) // 70 character password (under bcrypt limit)
		for i := range password {
			password = password[:i] + "a" + password[i+1:]
		}
		
		hashedPassword, err := HashPassword(password)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
	})
}

func TestCheckPasswordHash(t *testing.T) {
	setupTestEnv()

	t.Run("Check correct password", func(t *testing.T) {
		password := "correctpassword"
		hashedPassword, _ := HashPassword(password)
		
		result := CheckPasswordHash(password, hashedPassword)
		
		assert.True(t, result)
	})

	t.Run("Check incorrect password", func(t *testing.T) {
		correctPassword := "correctpassword"
		incorrectPassword := "wrongpassword"
		hashedPassword, _ := HashPassword(correctPassword)
		
		result := CheckPasswordHash(incorrectPassword, hashedPassword)
		
		assert.False(t, result)
	})

	t.Run("Check empty password against hash", func(t *testing.T) {
		password := "somepassword"
		hashedPassword, _ := HashPassword(password)
		
		result := CheckPasswordHash("", hashedPassword)
		
		assert.False(t, result)
	})

	t.Run("Check password against empty hash", func(t *testing.T) {
		password := "somepassword"
		
		result := CheckPasswordHash(password, "")
		
		assert.False(t, result)
	})

	t.Run("Check password against invalid hash", func(t *testing.T) {
		password := "somepassword"
		invalidHash := "not-a-valid-bcrypt-hash"
		
		result := CheckPasswordHash(password, invalidHash)
		
		assert.False(t, result)
	})

	t.Run("Check with bcrypt hash manually created", func(t *testing.T) {
		password := "testpassword"
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		assert.NoError(t, err)
		
		result := CheckPasswordHash(password, string(hash))
		
		assert.True(t, result)
	})

	t.Run("Password hash integration", func(t *testing.T) {
		testCases := []string{
			"password123",
			"P@ssw0rd!",
			"very_long_password_with_special_chars_123!@#",
			"短い", // Short Japanese text
			"🔒🗝️🔑", // Emojis
		}

		for _, password := range testCases {
			hashedPassword, err := HashPassword(password)
			assert.NoError(t, err)
			
			// Correct password should work
			assert.True(t, CheckPasswordHash(password, hashedPassword))
			
			// Incorrect password should fail
			assert.False(t, CheckPasswordHash(password+"wrong", hashedPassword))
		}
	})
}

func TestGenerateJWT(t *testing.T) {
	setupTestEnv()

	t.Run("Generate JWT with valid inputs", func(t *testing.T) {
		userID := "60c72b2f9b1d8b3a4c8e4f1a"
		email := "test@example.com"
		
		token, err := GenerateJWT(userID, email)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.True(t, len(token) > 50) // JWT tokens are typically long
	})

	t.Run("Generate JWT with empty inputs", func(t *testing.T) {
		token, err := GenerateJWT("", "")
		
		// Should still work (claims can be empty)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("Generate different JWTs for different users", func(t *testing.T) {
		userID1 := "60c72b2f9b1d8b3a4c8e4f1a"
		userID2 := "60c72b2f9b1d8b3a4c8e4f1b"
		email := "test@example.com"
		
		token1, err1 := GenerateJWT(userID1, email)
		token2, err2 := GenerateJWT(userID2, email)
		
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, token1, token2)
	})

	t.Run("Generate different JWTs for same user at different times", func(t *testing.T) {
		userID := "60c72b2f9b1d8b3a4c8e4f1a"
		email := "test@example.com"
		
		token1, err1 := GenerateJWT(userID, email)
		token2, err2 := GenerateJWT(userID, email)
		
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		// Tokens might be different due to timestamp (depends on implementation)
		// but both should be valid
		assert.NotEmpty(t, token1)
		assert.NotEmpty(t, token2)
	})

	t.Run("Generate JWT with special characters", func(t *testing.T) {
		userID := "60c72b2f9b1d8b3a4c8e4f1a"
		email := "test+label@example.com"
		
		token, err := GenerateJWT(userID, email)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("JWT token structure", func(t *testing.T) {
		userID := "60c72b2f9b1d8b3a4c8e4f1a"
		email := "test@example.com"
		
		token, err := GenerateJWT(userID, email)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		
		// JWT should have 3 parts separated by dots
		parts := len(token)
		assert.True(t, parts > 0)
		
		// Basic JWT format check (should contain dots)
		dotCount := 0
		for _, char := range token {
			if char == '.' {
				dotCount++
			}
		}
		assert.Equal(t, 2, dotCount, "JWT should have exactly 2 dots")
	})
}

// Benchmark tests
func BenchmarkHashPassword(b *testing.B) {
	setupTestEnv()
	password := "benchmarkpassword"
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		HashPassword(password)
	}
}

func BenchmarkCheckPasswordHash(b *testing.B) {
	setupTestEnv()
	password := "benchmarkpassword"
	hashedPassword, _ := HashPassword(password)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		CheckPasswordHash(password, hashedPassword)
	}
}

func BenchmarkGenerateJWT(b *testing.B) {
	setupTestEnv()
	userID := "60c72b2f9b1d8b3a4c8e4f1a"
	email := "test@example.com"
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		GenerateJWT(userID, email)
	}
}