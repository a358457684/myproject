package main

import (
	"context"
	"pp/common-golang/jws"
)

func main() {
	j := jws.NewJWT(nil, "0wsszP*VI4#)@Ekmq")
	c := context.Background()
	j.CreateToken(c, jws.TokenClaims{UserId: 128792327727632160})
}
