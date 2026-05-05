package general

import (
	"fmt"
	"strings"
)

func PrintBlock(title string, kv map[string]interface{}) {
	fmt.Println("================================")
	fmt.Printf(" %s\n", title)
	fmt.Println("================================")

	maxKey := 0
	for k := range kv {
		if len(k) > maxKey {
			maxKey = len(k)
		}
	}

	for k, v := range kv {
		str := fmt.Sprintf("%v", v)

		lines := strings.Split(str, "\n")
		if len(lines) == 1 {
			fmt.Printf(" %-*s : %s\n", maxKey, k, lines[0])
		} else {
			fmt.Printf(" %-*s : %s\n", maxKey, k, lines[0])
			for _, line := range lines[1:] {
				fmt.Printf(" %-*s   %s\n", maxKey, "", line)
			}
		}
	}

	fmt.Println()
}
