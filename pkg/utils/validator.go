package utils

import "unicode"

func IsValidUsername(username string) bool {
    for _, r := range username {
        if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
            return false
        }
    }
    return true
}