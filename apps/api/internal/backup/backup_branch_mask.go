package backup

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"slices"
	"strings"
)

type piiKind string

const (
	piiEmail piiKind = "email"
	piiName  piiKind = "name"
	piiPhone piiKind = "phone"
)

var (
	fieldWords = regexp.MustCompile(`[^a-z0-9]+`)
	camelWords = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

func sensitiveField(field string) piiKind {
	field = camelWords.ReplaceAllString(field, "${1}_${2}")
	words := fieldWords.Split(strings.ToLower(field), -1)
	for i := range words {
		switch strings.Join(words[i:], "") {
		case "email", "emails", "emailaddress", "emailaddresses":
			return piiEmail
		case "phone", "phones", "phonenumber", "phonenumbers", "mobilephone", "mobilenumber", "mobile", "telephone", "cellphone":
			return piiPhone
		case "fullname", "firstname", "lastname", "middlename", "givenname", "familyname", "surname", "forename", "displayname", "preferredname":
			return piiName
		case "name", "names":
			if i == 0 || slices.Contains([]string{"user", "customer", "contact", "person", "employee", "recipient", "billing", "shipping"}, words[i-1]) {
				return piiName
			}
		}
	}
	return ""
}

type branchMasker struct{ key string }

func newBranchMasker() branchMasker { return branchMasker{key: randomPassword()} }

func (m branchMasker) value(kind piiKind, value string) string {
	if value == "" {
		return value
	}
	for counter := byte(0); ; counter++ {
		hash := hmac.New(sha256.New, []byte(m.key))
		hash.Write([]byte(kind))
		hash.Write([]byte{counter})
		hash.Write([]byte(value))
		digest := hash.Sum(nil)
		var masked string
		switch kind {
		case piiEmail:
			masked = hex.EncodeToString(digest[:8]) + "@example.invalid"
		case piiName, piiPhone:
			var replacement strings.Builder
			i := 0
			for _, ch := range value {
				if kind == piiName && strings.ContainsRune(" -'.", ch) || kind == piiPhone && strings.ContainsRune(" +-().", ch) {
					replacement.WriteRune(ch)
					continue
				}
				b := digest[i%len(digest)]
				if kind == piiName {
					base := byte('a')
					if i == 0 || (ch >= 'A' && ch <= 'Z') {
						base = 'A'
					}
					replacement.WriteByte(base + b%26)
				} else if i == 0 && ch != '0' {
					replacement.WriteByte('1' + b%9)
				} else {
					replacement.WriteByte('0' + b%10)
				}
				i++
			}
			masked = replacement.String()
			if i == 0 {
				return masked
			}
		}
		if masked != value {
			return masked
		}
	}
}
