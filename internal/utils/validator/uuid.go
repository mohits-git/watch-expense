package validator

import "github.com/google/uuid"

func ValidateUUID(id string) bool {
  return uuid.Validate(id) == nil
}
