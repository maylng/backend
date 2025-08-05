package services

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/maylng/backend/internal/models"
)

type DNSValidationService struct{}

func NewDNSValidationService() *DNSValidationService {
	return &DNSValidationService{}
}

type DNSValidationResult struct {
	RecordType    string `json:"record_type"`
	RecordName    string `json:"record_name"`
	ExpectedValue string `json:"expected_value"`
	ActualValue   string `json:"actual_value,omitempty"`
	IsPresent     bool   `json:"is_present"`
	Error         string `json:"error,omitempty"`
}

type DomainDNSStatus struct {
	Domain            string                `json:"domain"`
	AllRecordsPresent bool                  `json:"all_records_present"`
	ValidationResults []DNSValidationResult `json:"validation_results"`
	CheckedAt         time.Time             `json:"checked_at"`
}

// ValidateDomainDNS checks if the required DNS records are present for a custom domain
func (s *DNSValidationService) ValidateDomainDNS(customDomain *models.CustomDomain) (*DomainDNSStatus, error) {
	status := &DomainDNSStatus{
		Domain:            customDomain.Domain,
		AllRecordsPresent: true,
		ValidationResults: []DNSValidationResult{},
		CheckedAt:         time.Now(),
	}

	// Check each DNS record
	for _, record := range customDomain.DNSRecords {
		result := s.checkDNSRecord(record)
		status.ValidationResults = append(status.ValidationResults, result)

		if !result.IsPresent {
			status.AllRecordsPresent = false
		}
	}

	return status, nil
}

// checkDNSRecord validates a single DNS record
func (s *DNSValidationService) checkDNSRecord(record models.DNSRecord) DNSValidationResult {
	result := DNSValidationResult{
		RecordType:    record.Type,
		RecordName:    record.Name,
		ExpectedValue: record.Value,
		IsPresent:     false,
	}

	switch strings.ToUpper(record.Type) {
	case "CNAME":
		result = s.checkCNAME(record)
	case "TXT":
		result = s.checkTXT(record)
	case "MX":
		result = s.checkMX(record)
	default:
		result.Error = fmt.Sprintf("Unsupported record type: %s", record.Type)
	}

	return result
}

// checkCNAME validates a CNAME record
func (s *DNSValidationService) checkCNAME(record models.DNSRecord) DNSValidationResult {
	result := DNSValidationResult{
		RecordType:    record.Type,
		RecordName:    record.Name,
		ExpectedValue: record.Value,
		IsPresent:     false,
	}

	// Look up CNAME record
	cname, err := net.LookupCNAME(record.Name)
	if err != nil {
		result.Error = fmt.Sprintf("DNS lookup failed: %v", err)
		return result
	}

	// Remove trailing dot from CNAME result
	cname = strings.TrimSuffix(cname, ".")
	expectedValue := strings.TrimSuffix(record.Value, ".")

	result.ActualValue = cname
	result.IsPresent = strings.EqualFold(cname, expectedValue)

	if !result.IsPresent {
		result.Error = fmt.Sprintf("CNAME mismatch: expected %s, got %s", expectedValue, cname)
	}

	return result
}

// checkTXT validates a TXT record
func (s *DNSValidationService) checkTXT(record models.DNSRecord) DNSValidationResult {
	result := DNSValidationResult{
		RecordType:    record.Type,
		RecordName:    record.Name,
		ExpectedValue: record.Value,
		IsPresent:     false,
	}

	// Look up TXT records
	txtRecords, err := net.LookupTXT(record.Name)
	if err != nil {
		result.Error = fmt.Sprintf("DNS lookup failed: %v", err)
		return result
	}

	// Check if expected value is in any of the TXT records
	for _, txt := range txtRecords {
		if strings.EqualFold(txt, record.Value) {
			result.IsPresent = true
			result.ActualValue = txt
			break
		}
	}

	if !result.IsPresent {
		result.ActualValue = strings.Join(txtRecords, "; ")
		result.Error = fmt.Sprintf("TXT record not found: expected %s, found %s", record.Value, result.ActualValue)
	}

	return result
}

// checkMX validates an MX record
func (s *DNSValidationService) checkMX(record models.DNSRecord) DNSValidationResult {
	result := DNSValidationResult{
		RecordType:    record.Type,
		RecordName:    record.Name,
		ExpectedValue: record.Value,
		IsPresent:     false,
	}

	// Look up MX records
	mxRecords, err := net.LookupMX(record.Name)
	if err != nil {
		result.Error = fmt.Sprintf("DNS lookup failed: %v", err)
		return result
	}

	// Check if expected value is in any of the MX records
	expectedHost := strings.TrimSuffix(record.Value, ".")
	for _, mx := range mxRecords {
		host := strings.TrimSuffix(mx.Host, ".")
		if strings.EqualFold(host, expectedHost) {
			result.IsPresent = true
			result.ActualValue = mx.Host
			break
		}
	}

	if !result.IsPresent {
		var hosts []string
		for _, mx := range mxRecords {
			hosts = append(hosts, mx.Host)
		}
		result.ActualValue = strings.Join(hosts, "; ")
		result.Error = fmt.Sprintf("MX record not found: expected %s, found %s", expectedHost, result.ActualValue)
	}

	return result
}

// GetDNSPropagationStatus provides a user-friendly summary of DNS propagation
func (s *DNSValidationService) GetDNSPropagationStatus(customDomain *models.CustomDomain) (string, error) {
	status, err := s.ValidateDomainDNS(customDomain)
	if err != nil {
		return "Error checking DNS", err
	}

	if status.AllRecordsPresent {
		return "All DNS records are properly configured", nil
	}

	missingCount := 0
	for _, result := range status.ValidationResults {
		if !result.IsPresent {
			missingCount++
		}
	}

	if missingCount == len(status.ValidationResults) {
		return "DNS records not found - please add the required records to your DNS", nil
	}

	return fmt.Sprintf("%d of %d DNS records are configured correctly",
		len(status.ValidationResults)-missingCount, len(status.ValidationResults)), nil
}
