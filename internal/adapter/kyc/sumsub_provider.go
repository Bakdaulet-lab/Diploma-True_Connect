package kyc

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/trueconnect/backend/internal/domain"
)

// SumsubProvider implements a real-world integration pattern for Sumsub KYC.
type SumsubProvider struct {
	apiToken string
	secret   string
	baseURL  string
	log      *slog.Logger
	client   *http.Client
}

// NewSumsubProvider creates a new instance of the Sumsub adapter.
func NewSumsubProvider(apiToken, secret string, log *slog.Logger) *SumsubProvider {
	return &SumsubProvider{
		apiToken: apiToken,
		secret:   secret,
		baseURL:  "https://api.sumsub.com",
		log:      log,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

// signRequest generates and adds Sumsub HMAC signatures to the request
func (p *SumsubProvider) signRequest(req *http.Request, body []byte) {
	ts := fmt.Sprintf("%d", time.Now().Unix())
	mac := hmac.New(sha256.New, []byte(p.secret))
	mac.Write([]byte(ts + req.Method + req.URL.RequestURI()))
	if len(body) > 0 {
		mac.Write(body)
	}
	signature := hex.EncodeToString(mac.Sum(nil))

	req.Header.Set("X-App-Token", p.apiToken)
	req.Header.Set("X-App-Access-Sig", signature)
	req.Header.Set("X-App-Access-Ts", ts)
}

// VerifyDocument creates an applicant in Sumsub and uploads their ID document.
// The actual verification level is updated asynchronously via Webhook.
func (p *SumsubProvider) VerifyDocument(ctx context.Context, userID uuid.UUID, documentKey string, documentData []byte, mimeType string) (domain.VerificationLevel, error) {
	p.log.Info("Sending document to external KYC provider (Sumsub)",
		slog.String("user_id", userID.String()),
		slog.Int("bytes", len(documentData)),
	)

	// In a real flow, you would:
	// 1. Create applicant with externalUserId = userID.String()
	applicantID, err := p.createApplicant(ctx, userID.String())
	if err != nil {
		p.log.Error("failed to create sumsub applicant", slog.String("error", err.Error()))
		return domain.VerificationNone, err
	}

	// 2. Add ID document to the applicant
	err = p.uploadDocument(ctx, applicantID, documentData, mimeType)
	if err != nil {
		p.log.Error("failed to upload sumsub document", slog.String("error", err.Error()))
		return domain.VerificationNone, err
	}

	// 3. Request applicant check (start verification process)
	err = p.requestApplicantCheck(ctx, applicantID)
	if err != nil {
		p.log.Error("failed to request applicant check", slog.String("error", err.Error()))
		return domain.VerificationNone, err
	}

	p.log.Info("Sumsub verification initiated, pending async webhook webhook response")
	// Return VerificationNone indicating it's still pending
	return domain.VerificationNone, nil
}

func (p *SumsubProvider) createApplicant(ctx context.Context, externalUserID string) (string, error) {
	payload := map[string]interface{}{
		"externalUserId": externalUserID,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/resources/applicants?levelName=basic-kyc-level", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	p.signRequest(req, body)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("sumsub create applicant failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Id string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Id, nil
}

func (p *SumsubProvider) uploadDocument(ctx context.Context, applicantID string, docData []byte, mimeType string) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add metadata
	metadata, _ := json.Marshal(map[string]string{"idDocType": "ID_CARD", "country": "USA"})
	metaField, _ := w.CreateFormField("metadata")
	metaField.Write(metadata)

	// Add file
	fw, _ := w.CreateFormFile("content", "document.jpg")
	fw.Write(docData)
	w.Close()

	uri := fmt.Sprintf("/resources/applicants/%s/info/idDoc", applicantID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+uri, &b)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	p.signRequest(req, b.Bytes())

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("sumsub upload document failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (p *SumsubProvider) requestApplicantCheck(ctx context.Context, applicantID string) error {
	payload := map[string]string{"reason": "User requested verification"}
	body, _ := json.Marshal(payload)

	uri := fmt.Sprintf("/resources/applicants/%s/status/pending", applicantID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+uri, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	p.signRequest(req, body)

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sumsub request check failed: status=%d", resp.StatusCode)
	}
	return nil
}
