package flows

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

// Terminal implements auth.UserAuthenticator prompting the terminal for
// input.
//
// This is only example implementation, you should not use it in your code.
// Copy it and modify to fit your needs.
type Terminal struct {
	PhoneNumber string // optional, will be prompted if empty
}

func (Terminal) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("signing up not implemented in Terminal")
}

func (Terminal) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

func (Terminal) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

func (a Terminal) Phone(_ context.Context) (string, error) {
	if a.PhoneNumber != "" {
		return a.PhoneNumber, nil
	}
	fmt.Print("Enter phone in international format (e.g. +1234567890): ")
	phone, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(phone), nil
}

func (Terminal) Password(_ context.Context) (string, error) {
	pwd, err := Ask("Enter 2FA password: ")
	if err != nil {
		return "", errors.New("Empty string!")
	}
	return strings.TrimSpace(pwd), nil
}

func Ask(prompt string) (string, error) {
	fmt.Print(prompt)
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	return input, err
}
