package licensing

import (
	"errors"
	"sync"
)

var (
	ErrLicenseExpired = errors.New("license expired")
	ErrLimitExceeded  = errors.New("license limit exceeded")
	ErrNoLicense      = errors.New("no active license found")
)

type LicenseManager struct {
	verifier       *Verifier
	currentLicense *License
	mu             sync.RWMutex
}

func NewLicenseManager(verifier *Verifier) *LicenseManager {
	return &LicenseManager{
		verifier: verifier,
	}
}

// LoadLicense loads a license from a string
func (m *LicenseManager) LoadLicense(licenseString string) error {
	license, err := m.verifier.Verify(licenseString)
	if err != nil {
		return err
	}

	if license.IsExpired() {
		return ErrLicenseExpired
	}

	m.mu.Lock()
	m.currentLicense = license
	m.mu.Unlock()

	return nil
}

// GetLicense returns the current license
func (m *LicenseManager) GetLicense() (*License, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentLicense == nil {
		return nil, ErrNoLicense
	}
	return m.currentLicense, nil
}

// CheckSaaSLimit checks if a SaaS usage limit has been exceeded
// usage is the current usage count for the period
func (m *LicenseManager) CheckSaaSLimit(metric string, usage int64) error {
	m.mu.RLock()
	license := m.currentLicense
	m.mu.RUnlock()

	if license == nil {
		return ErrNoLicense
	}

	if license.Type != LicenseTypeSaaS {
		return nil // Not applicable for On-Prem (or handle differently)
	}

	switch metric {
	case "api_requests":
		if license.Limits.MonthlyAPIRequests > 0 && usage >= license.Limits.MonthlyAPIRequests {
			return ErrLimitExceeded
		}
	case "pii_records":
		if license.Limits.PIIRecordsProcessed > 0 && usage >= license.Limits.PIIRecordsProcessed {
			return ErrLimitExceeded
		}
	}

	return nil
}

// CheckOnPremLimit checks if an On-Prem resource limit has been exceeded
func (m *LicenseManager) CheckOnPremLimit(metric string, currentCount int) error {
	m.mu.RLock()
	license := m.currentLicense
	m.mu.RUnlock()

	if license == nil {
		return ErrNoLicense
	}

	if license.Type != LicenseTypeOnPrem {
		return nil
	}

	switch metric {
	case "users":
		if license.Limits.MaxUsers > 0 && currentCount >= license.Limits.MaxUsers {
			return ErrLimitExceeded
		}
	case "domains":
		if license.Limits.MaxDomains > 0 && currentCount >= license.Limits.MaxDomains {
			return ErrLimitExceeded
		}
	}

	return nil
}
