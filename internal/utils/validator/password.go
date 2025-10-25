package validator

func ValidatePassword(password string) bool {
  if len(password) < 8 {
    return false
  }
  return true
}
