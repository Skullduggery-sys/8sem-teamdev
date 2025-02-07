package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"math"
	"math/big"
	"os"
	"sync"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

const (
	passwordCost = 14

	code2FADigits = 6
	smtpPort      = 587

	email2FAEnabledKey = "EMAIL_FA_ENABLED"
)

type MethodAfter2FA int

const (
	UserMethodSignUp MethodAfter2FA = iota
	UserMethodSignIn
	UserMethodResetPassword
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (int, error)
	GetByLogin(ctx context.Context, login string) (*model.User, error)
	ResetPassword(ctx context.Context, login, newPassword string) error
}

type AuthService struct {
	mx          sync.RWMutex
	sessions    map[string]*Session
	adminSecret string

	verifyCodes    map[string]string // user.email -> code
	verifyMetadata map[string]*VeryfyUserMetadata

	userRepo UserRepository
}

type VeryfyUserMetadata struct {
	User        *model.User
	Token       string
	NewPassword string

	Action MethodAfter2FA
}

func NewAuthService(userRepo UserRepository, adminSecret string) *AuthService {
	return &AuthService{
		mx:          sync.RWMutex{},
		sessions:    make(map[string]*Session),
		adminSecret: adminSecret,
		userRepo:    userRepo,

		verifyCodes:    make(map[string]string),
		verifyMetadata: make(map[string]*VeryfyUserMetadata),
	}
}

func (a *AuthService) GetUserID(token string) (int, error) {
	session, ok := a.sessions[token]
	if !ok {
		return 0, ErrNotFound
	}

	return session.UserID, nil
}

func (a *AuthService) GetUserTokenByAdmin(ctx context.Context, adminSecret, login string) (string, error) {
	if adminSecret != a.adminSecret {
		return "", ErrAdminIsNotAuthtorized
	}

	userInDB, err := a.userRepo.GetByLogin(ctx, login)
	if errors.Is(err, repository.ErrNotFound) {
		return "", ErrNotFound
	} else if err != nil {
		return "", err
	}

	session := NewSession(userInDB.ID, model.DefaultUser.String())

	a.mx.Lock()
	a.sessions[session.Token] = session
	a.mx.Unlock()

	return session.Token, nil
}

func (a *AuthService) SignUp(ctx context.Context, user *model.User) (int, error) {
	if user.Role == model.Admin.String() && user.AdminSecret != a.adminSecret {
		return 0, ErrAdminIsNotAuthtorized
	}

	if _, err := a.userRepo.GetByLogin(ctx, user.Login); err == nil {
		return 0, ErrLoginAlreadyExists
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), passwordCost)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrGeneratingHash, err)
	}

	user.Password = string(bytes)

	if verifyEnabled := os.Getenv(email2FAEnabledKey); verifyEnabled != "" {
		a.mx.Lock()
		defer a.mx.Unlock()

		code := generate2FACode()
		sendCodeInEmail(code)
		a.verifyCodes[user.Email] = code
		a.verifyMetadata[user.Email] = &VeryfyUserMetadata{
			User:   user,
			Action: UserMethodSignUp,
		}

		return 0, ErrWaiting2FA
	}

	return a.userRepo.Create(ctx, user)
}

func (a *AuthService) SignIn(ctx context.Context, user *model.User) (string, error) {
	userInDB, err := a.userRepo.GetByLogin(ctx, user.Login)
	if errors.Is(err, repository.ErrNotFound) {
		return "", ErrNotFound
	} else if err != nil {
		return "", err
	}

	hashedDBUser, err := bcrypt.GenerateFromPassword([]byte(user.Password), passwordCost)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrGeneratingHash, err)
	}

	if err = bcrypt.CompareHashAndPassword(hashedDBUser, []byte(user.Password)); err != nil {
		slog.Warn("password mismatch", "error", err)
		return "", ErrBadPassword
	} else if user.Role == model.Admin.String() && userInDB.Role != model.Admin.String() {
		return "", ErrAdminIsNotAuthtorized
	}

	if verifyEnabled := os.Getenv(email2FAEnabledKey); verifyEnabled != "" {
		a.mx.Lock()
		defer a.mx.Unlock()

		code := generate2FACode()
		sendCodeInEmail(code)
		a.verifyCodes[user.Email] = code
		a.verifyMetadata[user.Email] = &VeryfyUserMetadata{
			User:   user,
			Action: UserMethodSignIn,
		}

		return "", ErrWaiting2FA
	}

	session := NewSession(userInDB.ID, user.Role)

	a.mx.Lock()
	a.sessions[session.Token] = session
	a.mx.Unlock()

	return session.Token, nil
}

func (a *AuthService) SignOut(_ context.Context, token string) error {
	if _, ok := a.sessions[token]; !ok {
		return ErrNotFound
	}

	a.mx.Lock()
	delete(a.sessions, token)
	a.mx.Unlock()

	return nil
}

func (a *AuthService) ResetPassword(ctx context.Context, login, email, oldPassword, newPassword string) error {
	userInDB, err := a.userRepo.GetByLogin(ctx, login)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userInDB.Password), []byte(oldPassword)); err != nil {
		slog.Warn("password mismatch", "error", err)
		return ErrBadPassword
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), passwordCost)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGeneratingHash, err)
	}

	if verifyEnabled := os.Getenv(email2FAEnabledKey); verifyEnabled != "" {
		a.mx.Lock()
		defer a.mx.Unlock()

		code := generate2FACode()
		sendCodeInEmail(code)
		a.verifyCodes[email] = code
		a.verifyMetadata[email] = &VeryfyUserMetadata{
			User:        &model.User{Email: email, Login: login},
			NewPassword: string(newHash),
			Action:      UserMethodResetPassword,
		}

		return ErrWaiting2FA
	}

	return a.userRepo.ResetPassword(ctx, login, string(newHash))
}

func (a *AuthService) Handle2FA(ctx context.Context, email, code string) (*VeryfyUserMetadata, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if actualCode, ok := a.verifyCodes[email]; !ok {
		return nil, ErrNotFound
	} else if actualCode != code {
		return nil, ErrBadVeryfyCode
	}

	meta, ok := a.verifyMetadata[email]
	if !ok {
		return nil, ErrUserMetaNotFound
	}

	switch meta.Action {
	case UserMethodSignUp:
		id, err := a.userRepo.Create(ctx, meta.User)
		if err != nil {
			return nil, err
		}

		meta.User.ID = id
		return meta, nil
	case UserMethodSignIn:
		session := NewSession(meta.User.ID, meta.User.Role)
		a.sessions[session.Token] = session
		meta.Token = session.Token
		return meta, nil
	case UserMethodResetPassword:
		if err := a.userRepo.ResetPassword(ctx, meta.User.Login, meta.NewPassword); err != nil {
			return nil, err
		}

		return meta, nil
	}

	return nil, ErrVerifyActionNotSaved
}

func (a *AuthService) ClearVerifyEmail(email string) {
	a.mx.Lock()
	defer a.mx.Unlock()

	delete(a.verifyMetadata, email)
	delete(a.verifyCodes, email)
	slog.Info("cleared verify metadata & code", "user_email", email)
}

func sendCodeInEmail(code string) {
	senderEmail := os.Getenv("SENDER_EMAIL_ADDRESS")
	senderPassword := os.Getenv("SENDER_EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_SERVER")
	recipientEmail := os.Getenv("RECIPIENT_EMAIL_ADDRESS")
	m := gomail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", recipientEmail)
	m.SetHeader("Subject", "[7sem-testing]: Your Verification Code")
	m.SetBody("text/plain", fmt.Sprintf("Your verification code is: %s", code))
	d := gomail.NewDialer(smtpHost, smtpPort, senderEmail, senderPassword)
	if err := d.DialAndSend(m); err != nil {
		log.Fatalf("Error sending verification code: %v", err)
	}
}

func generate2FACode() string {
	if ok := os.Getenv("E2E_TEST"); ok != "" {
		return "228228"
	}

	bi, err := rand.Int(
		rand.Reader,
		big.NewInt(int64(math.Pow(10, float64(code2FADigits)))),
	)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%0*d", code2FADigits, bi)
}
