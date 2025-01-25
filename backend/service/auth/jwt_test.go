package auth

import "testing"

func TestCreateJWT(t *testing.T){
	secret := []byte("secret")

	token, err := CreateJWT(secret, 1)
	if err != nil {
		t.Errorf("error creating jwt %v", err)
	}

	if len(token) == 0 {
		t.Error("expected token to be not empty")
	}

}