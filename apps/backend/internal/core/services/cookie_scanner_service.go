package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"pixpivot/arc/internal/dto"
	"pixpivot/arc/internal/models"
	"pixpivot/arc/internal/storage/repository"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/datatypes"
)

type CookieScannerService struct {
	repo *repository.CookieRepository
}

func NewCookieScannerService(repo *repository.CookieRepository) *CookieScannerService {
	return &CookieScannerService{repo: repo}
}

type ScanResult struct {
	Cookies      []dto.DetectedCookie `json:"cookies"`
	CookiesFound int                  `json:"cookiesFound"`
	NewCookies   int                  `json:"newCookies"`
	ScanDuration int                  `json:"scanDuration"`
	PagesScanned int                  `json:"pagesScanned"`
}

func (s *CookieScannerService) ScanWebsite(tenantID uuid.UUID, targetURL string) (*models.CookieScan, error) {
	// Create initial scan record
	scan := &models.CookieScan{
		ID:       uuid.New(),
		TenantID: tenantID,
		URL:      targetURL,
		Status:   models.CookieScanStatusPending,
	}

	if err := s.repo.CreateScan(scan); err != nil {
		return nil, fmt.Errorf("failed to create scan record: %w", err)
	}

	// Start scanning in background
	go s.performScan(scan)

	return scan, nil
}

func (s *CookieScannerService) performScan(scan *models.CookieScan) {
	startTime := time.Now()

	// Update status to running
	scan.Status = models.CookieScanStatusRunning
	s.repo.UpdateScan(scan)

	// Perform the actual scan with crawling
	result, err := s.scanWithCrawling(scan.URL)
	if err != nil {
		log.Error().Err(err).Str("url", scan.URL).Msg("Scan failed")
		scan.Status = models.CookieScanStatusFailed
		scan.ErrorMessage = err.Error()
		s.repo.UpdateScan(scan)
		return
	}

	// Compare with existing cookies to find new ones
	newCookiesCount, err := s.compareWithExisting(scan.TenantID, result.Cookies)
	if err != nil {
		log.Error().Err(err).Msg("Failed to compare cookies")
		scan.Status = models.CookieScanStatusFailed
		scan.ErrorMessage = fmt.Sprintf("Failed to compare cookies: %v", err)
		s.repo.UpdateScan(scan)
		return
	}

	// Update scan with results
	resultsJSON, _ := json.Marshal(result)
	scan.Results = datatypes.JSON(resultsJSON)
	scan.CookiesFound = result.CookiesFound
	scan.NewCookies = newCookiesCount
	scan.ScanDuration = int(time.Since(startTime).Milliseconds())
	scan.Status = models.CookieScanStatusCompleted

	s.repo.UpdateScan(scan)
}

func (s *CookieScannerService) scanWithCrawling(startURL string) (*ScanResult, error) {
	// Parse base domain to ensure we stay within the site
	parsedURL, err := url.Parse(startURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	baseDomain := parsedURL.Hostname()

	// URLs to visit (queue)
	urlsToVisit := []string{startURL}
	visitedURLs := make(map[string]bool)

	// Limit pages to scan to avoid taking too long
	maxPages := 5

	var allCookies []dto.DetectedCookie
	cookieMap := make(map[string]bool) // To deduplicate cookies by name+domain

	// Create a context for the entire session to share cookies/cache
	// Use a custom allocator to disable headless mode if needed for debugging, but usually headless is fine
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		chromedp.WindowSize(1920, 1080),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set a global timeout for the entire scan
	ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	pagesScanned := 0

	for len(urlsToVisit) > 0 && pagesScanned < maxPages {
		currentURL := urlsToVisit[0]
		urlsToVisit = urlsToVisit[1:]

		if visitedURLs[currentURL] {
			continue
		}
		visitedURLs[currentURL] = true
		pagesScanned++

		log.Info().Str("url", currentURL).Msg("Scanning page")

		// Scan the page
		pageCookies, internalLinks, err := s.scanPage(ctx, currentURL, baseDomain)
		if err != nil {
			log.Warn().Err(err).Str("url", currentURL).Msg("Failed to scan page, skipping")
			continue
		}

		// Add found cookies
		for _, c := range pageCookies {
			key := c.Name + "@" + c.Domain
			if !cookieMap[key] {
				cookieMap[key] = true
				allCookies = append(allCookies, c)
			}
		}

		// Add new internal links to queue
		for _, link := range internalLinks {
			if !visitedURLs[link] {
				// Simple check to avoid adding duplicates to queue (not perfect but helpful)
				isQueued := false
				for _, q := range urlsToVisit {
					if q == link {
						isQueued = true
						break
					}
				}
				if !isQueued {
					urlsToVisit = append(urlsToVisit, link)
				}
			}
		}
	}

	return &ScanResult{
		Cookies:      allCookies,
		CookiesFound: len(allCookies),
		PagesScanned: pagesScanned,
	}, nil
}

func (s *CookieScannerService) scanPage(ctx context.Context, targetURL, baseDomain string) ([]dto.DetectedCookie, []string, error) {
	// Create a timeout for this specific page
	pageCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	var cookies []*network.Cookie
	var links []string

	// Run chromedp tasks
	err := chromedp.Run(pageCtx,
		network.Enable(),
		chromedp.Navigate(targetURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Wait for dynamic content

		// Attempt to click "Accept All" or similar buttons if found (simple heuristic)
		// This is a best-effort attempt to trigger cookies
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Selectors for common consent buttons
			selectors := []string{
				"button[id*='accept']", "button[class*='accept']",
				"a[id*='accept']", "a[class*='accept']",
				"button:contains('Accept')", "button:contains('Agree')",
			}
			for _, sel := range selectors {
				// We don't want to fail if not found, just try to click
				_ = chromedp.Click(sel, chromedp.ByQuery).Do(ctx)
			}
			return nil
		}),
		chromedp.Sleep(1*time.Second), // Wait after potential click

		// Get cookies
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			cookies, err = network.GetCookies().Do(ctx)
			return err
		}),

		// Extract internal links
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).map(a => a.href)`, &links),
	)

	if err != nil {
		return nil, nil, err
	}

	// Filter and process links
	var internalLinks []string
	for _, link := range links {
		// Basic filtering
		if link == "" || strings.HasPrefix(link, "javascript:") || strings.HasPrefix(link, "mailto:") || strings.HasPrefix(link, "#") {
			continue
		}

		// Parse to check domain
		u, err := url.Parse(link)
		if err != nil {
			continue
		}

		// Check if internal
		if u.Hostname() == baseDomain || strings.HasSuffix(u.Hostname(), "."+baseDomain) {
			// Normalize URL (remove fragment)
			u.Fragment = ""
			internalLinks = append(internalLinks, u.String())
		}
	}

	// Convert cookies
	detectedCookies := make([]dto.DetectedCookie, 0, len(cookies))
	for _, cookie := range cookies {
		detected := s.convertCookie(cookie, baseDomain)
		detectedCookies = append(detectedCookies, detected)
	}

	return detectedCookies, internalLinks, nil
}

func (s *CookieScannerService) convertCookie(cookie *network.Cookie, targetDomain string) dto.DetectedCookie {
	// Determine if it's first-party
	isFirstParty := s.isFirstPartyCookie(cookie.Domain, targetDomain)

	// Calculate expiry days
	expiryDays := 0
	if cookie.Expires > 0 {
		expiryTime := time.Unix(int64(cookie.Expires), 0)
		expiryDays = int(time.Until(expiryTime).Hours() / 24)
	}

	// Categorize the cookie
	category := s.categorizeCookie(cookie.Name, cookie.Domain)

	// Get purpose and provider
	purpose, provider := s.getCookiePurposeAndProvider(cookie.Name, cookie.Domain, category)

	return dto.DetectedCookie{
		Name:         cookie.Name,
		Domain:       cookie.Domain,
		Path:         cookie.Path,
		Value:        cookie.Value,
		ExpiryDays:   expiryDays,
		IsFirstParty: isFirstParty,
		IsSecure:     cookie.Secure,
		IsHttpOnly:   cookie.HTTPOnly,
		SameSite:     string(cookie.SameSite),
		Category:     category,
		Purpose:      purpose,
		Provider:     provider,
	}
}

func (s *CookieScannerService) categorizeCookie(name, domain string) string {
	name = strings.ToLower(name)
	domain = strings.ToLower(domain)

	// Necessary cookies patterns
	necessaryPatterns := []string{
		"session", "sess", "csrf", "xsrf", "auth", "login", "security",
		"phpsessid", "jsessionid", "asp.net_sessionid", "connect.sid",
		"cf_bm", "__cf_bm", // Cloudflare
	}

	// Analytics cookies patterns
	analyticsPatterns := []string{
		"_ga", "_gid", "_gat", "_gtag", "_utm", "analytics", "_dc_gtm",
		"_hjid", "_hjIncludedInSample", "_hjAbsoluteSessionInProgress",
		"amplitude", "mixpanel", "segment", "_fbp", "_fbc",
		"mp_", "ajs_", // Mixpanel, Segment
	}

	// Marketing cookies patterns
	marketingPatterns := []string{
		"_gcl", "ads", "doubleclick", "googlesyndication", "facebook",
		"twitter", "linkedin", "pinterest", "instagram", "tiktok",
		"criteo", "outbrain", "taboola", "adsystem", "adnxs",
		"uuid2", "sess", "id", // Generic ad tech often uses short names, need to be careful
		"mc_", "ma_", // Mailchimp etc
	}

	// Check patterns
	for _, pattern := range necessaryPatterns {
		if strings.Contains(name, pattern) {
			return models.CookieCategoryNecessary
		}
	}

	for _, pattern := range analyticsPatterns {
		if strings.Contains(name, pattern) {
			return models.CookieCategoryAnalytics
		}
	}

	for _, pattern := range marketingPatterns {
		if strings.Contains(name, pattern) {
			return models.CookieCategoryMarketing
		}
	}

	// Default to functional if no specific pattern matches
	return models.CookieCategoryFunctional
}

func (s *CookieScannerService) getCookiePurposeAndProvider(name, domain, category string) (string, string) {
	name = strings.ToLower(name)
	domain = strings.ToLower(domain)

	// Common cookie purposes and providers
	// Expanded database
	cookieDatabase := map[string]struct {
		purpose  string
		provider string
	}{
		"_ga":              {"Google Analytics tracking", "Google"},
		"_gid":             {"Google Analytics session tracking", "Google"},
		"_gat":             {"Google Analytics throttling", "Google"},
		"_fbp":             {"Facebook Pixel tracking", "Facebook"},
		"_fbc":             {"Facebook click tracking", "Facebook"},
		"_hjid":            {"Hotjar user identification", "Hotjar"},
		"_utm":             {"Campaign tracking", "Various"},
		"session":          {"User session management", "Website"},
		"csrf":             {"Cross-site request forgery protection", "Website"},
		"cf_bm":            {"Cloudflare bot management", "Cloudflare"},
		"__cf_bm":          {"Cloudflare bot management", "Cloudflare"},
		"ajs_user_id":      {"Segment user identification", "Segment"},
		"ajs_anonymous_id": {"Segment anonymous identification", "Segment"},
		"mp_":              {"Mixpanel analytics", "Mixpanel"},
	}

	// Check for exact matches first
	if info, exists := cookieDatabase[name]; exists {
		return info.purpose, info.provider
	}

	// Check for partial matches
	for cookieName, info := range cookieDatabase {
		if strings.Contains(name, cookieName) {
			return info.purpose, info.provider
		}
	}

	// Determine provider from domain
	provider := s.getProviderFromDomain(domain)

	// Generate purpose based on category
	purpose := s.generatePurposeFromCategory(category)

	return purpose, provider
}

func (s *CookieScannerService) getProviderFromDomain(domain string) string {
	domain = strings.ToLower(domain)

	providers := map[string]string{
		"google":            "Google",
		"facebook":          "Facebook",
		"twitter":           "Twitter",
		"linkedin":          "LinkedIn",
		"youtube":           "Google",
		"doubleclick":       "Google",
		"googlesyndication": "Google",
		"hotjar":            "Hotjar",
		"mixpanel":          "Mixpanel",
		"amplitude":         "Amplitude",
		"segment":           "Segment",
		"criteo":            "Criteo",
		"outbrain":          "Outbrain",
		"taboola":           "Taboola",
		"cloudflare":        "Cloudflare",
		"amazon":            "Amazon",
		"aws":               "Amazon Web Services",
	}

	for domainPattern, provider := range providers {
		if strings.Contains(domain, domainPattern) {
			return provider
		}
	}

	return "Unknown"
}

func (s *CookieScannerService) generatePurposeFromCategory(category string) string {
	switch category {
	case models.CookieCategoryNecessary:
		return "Essential for website functionality and security"
	case models.CookieCategoryFunctional:
		return "Enhance website functionality and user experience"
	case models.CookieCategoryAnalytics:
		return "Analyze website usage and performance"
	case models.CookieCategoryMarketing:
		return "Deliver targeted advertising and marketing content"
	default:
		return "Website functionality"
	}
}

func (s *CookieScannerService) isFirstPartyCookie(cookieDomain, targetDomain string) bool {
	// Remove leading dots
	cookieDomain = strings.TrimPrefix(cookieDomain, ".")
	targetDomain = strings.TrimPrefix(targetDomain, ".")

	// Check if cookie domain matches or is a subdomain of target domain
	return cookieDomain == targetDomain || strings.HasSuffix(targetDomain, "."+cookieDomain)
}

func (s *CookieScannerService) extractDomain(targetURL string) string {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return ""
	}
	return parsedURL.Hostname()
}

func (s *CookieScannerService) compareWithExisting(tenantID uuid.UUID, detectedCookies []dto.DetectedCookie) (int, error) {
	newCookiesCount := 0

	for i := range detectedCookies {
		cookie := &detectedCookies[i]

		// Check if cookie already exists
		existing, err := s.repo.FindByNameAndDomain(tenantID, cookie.Name, cookie.Domain)
		if err != nil {
			// Cookie doesn't exist, mark as new
			cookie.IsNew = true
			newCookiesCount++

			// Create new cookie record
			newCookie := &models.Cookie{
				ID:            uuid.New(),
				TenantID:      tenantID,
				Name:          cookie.Name,
				Domain:        cookie.Domain,
				Path:          cookie.Path,
				Category:      cookie.Category,
				Purpose:       cookie.Purpose,
				Provider:      cookie.Provider,
				ExpiryDays:    cookie.ExpiryDays,
				IsFirstParty:  cookie.IsFirstParty,
				IsSecure:      cookie.IsSecure,
				IsHttpOnly:    cookie.IsHttpOnly,
				SameSite:      cookie.SameSite,
				Description:   cookie.Purpose,
				DataCollected: s.getDataCollectedDescription(cookie.Category),
				IsActive:      true,
			}

			s.repo.Create(newCookie)
		} else {
			// Cookie exists, check if category has changed
			if existing.Category != cookie.Category {
				existing.Category = cookie.Category
				existing.Purpose = cookie.Purpose
				existing.Provider = cookie.Provider
				s.repo.Update(existing)
			}
		}
	}

	return newCookiesCount, nil
}

func (s *CookieScannerService) getDataCollectedDescription(category string) string {
	switch category {
	case models.CookieCategoryNecessary:
		return "Session identifiers, authentication tokens, security tokens"
	case models.CookieCategoryFunctional:
		return "User preferences, language settings, feature toggles"
	case models.CookieCategoryAnalytics:
		return "Page views, user interactions, performance metrics, user journey data"
	case models.CookieCategoryMarketing:
		return "User interests, demographic data, ad interaction data, conversion tracking"
	default:
		return "Various user interaction data"
	}
}

func (s *CookieScannerService) GetScanResults(scanID, tenantID uuid.UUID) (*models.CookieScan, error) {
	return s.repo.GetScanByID(scanID, tenantID)
}

func (s *CookieScannerService) GetScanHistory(tenantID uuid.UUID, limit int) ([]*models.CookieScan, error) {
	return s.repo.ListScansByTenant(tenantID, limit)
}

func (s *CookieScannerService) DeleteScan(scanID, tenantID uuid.UUID) error {
	return s.repo.DeleteScan(scanID, tenantID)
}
