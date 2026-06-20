package cli

import "strconv"

// plural renders a natural-English count: "1 object" vs "3 objects".
func plural(n int, singular, pluralForm string) string {
	if n == 1 {
		return strconv.Itoa(n) + " " + singular
	}
	return strconv.Itoa(n) + " " + pluralForm
}
