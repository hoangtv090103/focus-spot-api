package hash

import "golang.org/x/crypto/bcrypt"

// GenerateHash creates a bcrypt hash of the password
func GenerateHash(password string) (string, error) {
    hashBytes, err :=bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    
    return string(hashBytes), nil
}

// VerifyHash checks if the password matches the hashed password
func VerifyHash(password, hashedPassword string) bool {
    bPassword := []byte(password)
    bHashedPassword := []byte(hashedPassword)
    err := bcrypt.CompareHashAndPassword(bHashedPassword, bPassword)
    return err == nil
}