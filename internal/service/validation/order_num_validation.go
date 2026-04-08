package validation

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func LuhnNumVerify(num string) error {

	if len(num) == 0 {
		return errors.New("empty num")
	}

	digits := strings.Split(num, "")

	mod := len(num) % 2
	sum := 0

	for i, f := range digits {

		digit, err := strconv.Atoi(f)

		if err != nil {
			return err
		}

		if i%2 == mod {
			doubled := digit * 2

			if doubled > 9 {
				doubled -= 9
			}

			sum += doubled

		} else {
			sum += digit
		}
	}

	if sum%10 != 0 {
		return fmt.Errorf("invalid code")
	}

	return nil
}
