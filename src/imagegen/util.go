package imagegen

import "strings"

func ToSnakecase(name string) string {
	lower := strings.TrimSpace(strings.ToLower(name))
	clean := strings.ReplaceAll(strings.ReplaceAll(lower, "(", ""), ")", "")
	elems := strings.Split(clean, " ")
	return strings.Join(elems, "_")
}
