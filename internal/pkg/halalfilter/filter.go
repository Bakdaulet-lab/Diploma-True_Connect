package halalfilter

import "strings"

// blockedWords triggers a hard block — message is rejected.
var blockedWords = []string{
	// explicit sexual content
	"секс", "порно", "интим", "голая", "голый", "эротик",
	"sex", "porn", "nude", "naked", "xxx", "erotic",
	// solicitation
	"проституция", "эскорт", "escort",
}

// warnedWords flag the message as toxic but do not reject outright.
var warnedWords = []string{
	// scam signals
	"scam", "мошенничество", "кидалово",
	// harassment / pressure
	"abuse", "harassment",
	// original test placeholders (keep for backwards-compat with tests)
	"badword",
}

// CheckMessage inspects text against the halal content filter.
// blocked=true → reject the message entirely.
// warned=true  → flag as toxic, apply trust penalty, but allow delivery.
func CheckMessage(text string) (blocked bool, warned bool) {
	lower := strings.ToLower(text)
	for _, w := range blockedWords {
		if strings.Contains(lower, w) {
			return true, false
		}
	}
	for _, w := range warnedWords {
		if strings.Contains(lower, w) {
			return false, true
		}
	}
	return false, false
}
