package validation

import (
	"fmt"
	"strconv"
	"strings"
)

func LuhnNumVerify(num string) error {
	digits := strings.Split(num, "")

	mod := len(num) % 2
	sum := 0

	for i, f := range digits {

		digit, err := strconv.Atoi(f)

		if err != nil {
			return err
		}

		if i%2 == mod {
			sum += digit * 2 % 9
		} else {
			sum += digit
		}
	}

	if sum%10 != 0 {
		return fmt.Errorf("invalid code")
	}

	return nil
}
