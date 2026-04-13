package service

import (
	"context"
	"strings"

	"github.com/trueconnect/backend/internal/pkg/crypto"
	"github.com/trueconnect/backend/internal/repository"
)

type AdminService struct {
	adminRepo     repository.AdminRepository
	encryptionKey []byte
}

func NewAdminService(adminRepo repository.AdminRepository, encryptionKey []byte) *AdminService {
	return &AdminService{adminRepo: adminRepo, encryptionKey: encryptionKey}
}

func (s *AdminService) GetDashboardStats(ctx context.Context) (*repository.AdminDashboardStats, error) {
	return s.adminRepo.GetDashboardStats(ctx)
}

func (s *AdminService) SearchUsers(ctx context.Context, searchQuery string, limit, offset int) ([]*repository.AdminUserRow, error) {
	var phoneHash []byte
	nameQuery := ""

	// Simple heuristic: if query contains numbers, treat it as a phone search
	if searchQuery != "" && strings.ContainsAny(searchQuery, "0123456789+") {
		phoneHash = crypto.SHA256Hash([]byte(searchQuery))
	} else if searchQuery != "" {
		nameQuery = searchQuery
	}

	users, err := s.adminRepo.SearchUsers(ctx, phoneHash, nameQuery, limit, offset)
	if err != nil {
		return nil, err
	}

	// Decrypt phone and email for admin view
	for _, u := range users {
		if decryptedPhone, err := crypto.Decrypt(u.PhoneEncrypted, s.encryptionKey); err == nil {
			u.Phone = string(decryptedPhone)
		}
		if u.EmailEncrypted != nil {
			if decryptedEmail, err := crypto.Decrypt(u.EmailEncrypted, s.encryptionKey); err == nil {
				u.Email = string(decryptedEmail)
			}
		}
	}

	return users, nil
}
