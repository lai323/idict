package wordset

import (
	"fmt"
	"testing"
)

func TestValidword(t *testing.T) {
	fmt.Println(validword.MatchString("abc"))
	fmt.Println(validword.MatchString("1abc"))
	fmt.Println(validword.MatchString("abc1"))
	fmt.Println(validword.MatchString("ab c"))
	fmt.Println(validword.MatchString("ab嘿"))
}
