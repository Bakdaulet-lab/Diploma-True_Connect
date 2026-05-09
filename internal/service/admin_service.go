package service

import (
	"context"
	"fmt"
	"github.com/trueconnect/backend/internal/domain"
	"strings"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/pkg/crypto"
	"github.com/trueconnect/backend/internal/repository"
)

type AdminService struct {
	adminRepo     repository.AdminRepository
	userRepo      repository.UserRepository
	reputeSvc     *ReputationService
	encryptionKey []byte
}

func NewAdminService(adminRepo repository.AdminRepository, userRepo repository.UserRepository, reputeSvc *ReputationService, encryptionKey []byte) *AdminService {
	return &AdminService{adminRepo: adminRepo, userRepo: userRepo, reputeSvc: reputeSvc, encryptionKey: encryptionKey}
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

func (s *AdminService) GetPendingKYC(ctx context.Context, limit, offset int) ([]*repository.PendingKYC, error) {
	return s.adminRepo.GetPendingKYC(ctx, limit, offset)
}

func (s *AdminService) GetKYCDocument(ctx context.Context, kycID uuid.UUID) (string, error) {
	return s.adminRepo.GetKYCDocument(ctx, kycID)
}

func (s *AdminService) UpdateKYCStatus(ctx context.Context, kycID uuid.UUID, status string) (uuid.UUID, error) {
	return s.adminRepo.UpdateKYCStatus(ctx, kycID, status)
}

func (s *AdminService) GetPendingSybilClusters(ctx context.Context, limit, offset int) ([]*repository.SybilClusterRow, error) {
	return s.adminRepo.GetPendingSybilClusters(ctx, limit, offset)
}

func (s *AdminService) ResolveSybilCluster(ctx context.Context, clusterID uuid.UUID, action string) error {
	var status string
	if action == "ban" {
		status = "banned"
	} else if action == "dismiss" {
		status = "dismissed"
	} else {
		return fmt.Errorf("invalid action: %s", action)
	}

	suspects, err := s.adminRepo.ResolveSybilCluster(ctx, clusterID, status)
	if err != nil {
		return err
	}

	// Based on the action, update the suspect users
	for _, uid := range suspects {
		if action == "ban" {
			// Update status to banned
			_ = s.userRepo.UpdateTrustStatus(ctx, uid, domain.TrustStatusBanned)

			// Hard-delete their Graph influence, forces depth-1 recalculation
			_ = s.reputeSvc.PurgeUserGraphInfluence(ctx, uid)
		} else {
			// Revert back to normal
			_ = s.userRepo.UpdateTrustStatus(ctx, uid, domain.TrustStatusNormal)
		}
	}

	return nil
}
