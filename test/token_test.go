package test

import (
	"douyin/tools"
	"fmt"
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestTokenValid(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJUb2tlbktleSI6ImxvZ2luX3Rva2VuczpiZjI0OWQxNi1kNjVmLTExZWMtOTI3Yi04YzhjYWE0NGZkYmQifQ.B_suKZqp11NPLBISm80CeAPPFP_-KXguR0S7h3AheVg"
	key, err := tools.JwtParseTokenKey(token)
	if err != nil {
		return
	}
	fmt.Println(key)
	err = tools.VeifyToken(token)
	if err != nil {
		fmt.Println(err)
	}
	assert.Equal(t, err, nil)
}
