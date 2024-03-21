package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"vbook/internal/repository"
	"vbook/internal/service/sms"
)

var ErrCodeSendTooMany = repository.ErrCodeVerifyTooMany

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}
type codeService struct {
	repo repository.CodeRepository
	sms  sms.Service
}

func NewCodeService(repo repository.CodeRepository, smsCs sms.Service) CodeService {
	return &codeService{
		repo: repo,
		sms:  smsCs,
	}
}

func (c *codeService) Send(ctx context.Context, biz string, phone string) error {
	code := c.generate()
	err := c.repo.Set(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	const codeTplId = "19370165"
	return c.sms.Send(ctx, codeTplId, []string{code}, phone)
}

func (c *codeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	ok, err := c.repo.Verify(ctx, biz, phone, inputCode)
	if errors.Is(err, ErrCodeSendTooMany) {
		return false, err
	}
	return ok, err
}
func (c *codeService) generate() string {
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}
