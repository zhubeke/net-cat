package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func IsValidName(name string) bool {
	name = strings.TrimSpace(name)
	return name != "" && !strings.Contains(name, "\n") && !strings.Contains(name, "\r")
}
