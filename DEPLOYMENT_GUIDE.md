# Backend Deployment Guide - License Fix & Auth System

## Summary of Changes

This deployment includes critical fixes for the authentication system:

1. **License Middleware Fix** - Allows auth endpoints to bypass licensing
2. **Email Service Enhancement** - Connection validation, retry logic, timeouts
3. **Atomic Email Transactions** - Signup rollback if email fails
4. **SSO Handler Rewrite** - Strict mode enforcement (blocks non-existent users)
5. **Login Verification** - Email verification required before login

---

## Pre-Deployment Checklist

- [ ] Backup production database
- [ ] Review all changed files (see below)
- [ ] Update SMTP credentials in production `.env`
- [ ] Ensure Docker/K8s deployment pipeline is ready

---

## Changed Files

### Critical Files
1. `cmd/server/main.go` - **License middleware path fix** (Line 165)
2. `internal/api/handlers/sso_handler.go` - **Complete rewrite** (strict mode)
3. `internal/api/handlers/signup_handler.go` - **Atomic email transactions**
4. `internal/api/handlers/auth_handler.go` - **Login verification checks**
5. `internal/core/services/email_service.go` - **Retry logic & validation**

### View Changes
```powershell
# Review the key changes
git diff cmd/server/main.go
git diff internal/api/handlers/sso_handler.go
git diff internal/api/handlers/signup_handler.go
git diff internal/api/handlers/auth_handler.go
git diff internal/core/services/email_service.go
```

---

## Deployment Steps

### Option 1: Docker Deployment

```bash
# Navigate to backend directory
cd d:\arc-demo\apps\backend

# Build new Docker image
docker build -t arc-backend:v2.0-auth-fix .

# Tag for registry (replace with your registry)
docker tag arc-backend:v2.0-auth-fix your-registry.com/arc-backend:v2.0-auth-fix

# Push to registry
docker push your-registry.com/arc-backend:v2.0-auth-fix

# Update deployment (if using K8s)
kubectl set image deployment/arc-backend arc-backend=your-registry.com/arc-backend:v2.0-auth-fix

# Or restart with docker-compose
docker-compose down
docker-compose up -d
```

### Option 2: Direct Server Deployment

```bash
# SSH to production server
ssh your-server

# Pull latest code
cd /path/to/arc-demo/apps/backend
git pull origin main

# Rebuild application
go build -o bin/server cmd/server/main.go

# Restart service
sudo systemctl restart arc-backend
# OR
pm2 restart arc-backend
```

### Option 3: Kubernetes Deployment

```bash
# Apply updated deployment
kubectl apply -f d:\arc-demo\k8s\backend-deployment.yaml

# Wait for rollout
kubectl rollout status deployment/arc-backend

# Verify pods are running
kubectl get pods -l app=arc-backend
```

---

## Post-Deployment Verification

### 1. Health Check
```powershell
curl https://arc-demo.thepixpivot.com/health
# Expected: "OK"
```

### 2. License Bypass Verification
```powershell
# Run automated test suite
cd d:\arc-demo
.\test-backend-production.ps1
```

**Expected Results**:
- ✅ Test 1: Fiduciary signup attempts (503 if SMTP not configured, 200 if configured)
- ✅ Test 2: Login returns 401 (invalid credentials) NOT 402 (license required)
- ✅ Test 3: Duplicate signup returns 400 NOT 402
- ✅ Test 4: Invalid credentials returns 401 NOT 402
- ✅ Test 5: User signup attempts (503 or 200) NOT 402

### 3. Manual API Test

```powershell
# Test signup endpoint is accessible
$response = Invoke-WebRequest -Uri "https://arc-demo.thepixpivot.com/api/v1/auth/fiduciary/signup" `
  -Method POST `
  -Headers @{"Content-Type"="application/json"} `
  -Body '{"email":"test@example.com","firstName":"Test","lastName":"User","password":"Test123!","confirmPassword":"Test123!","organization":{"name":"Test Org","industry":"tech","companySize":"1-10","address":"123 St","country":"US","email":"org@test.com","phone":"+1234567890"}}' `
  -ErrorAction SilentlyContinue

$response.StatusCode
# Expected: 503 (SMTP not configured) or 201 (success)
# NOT EXPECTED: 402 (license error)
```

---

## SMTP Configuration

### Update Production Environment Variables

**File**: `/path/to/backend/.env.production` or Kubernetes Secret

```bash
# Gmail Example (recommended for testing)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-actual-email@gmail.com
SMTP_PASS=your-app-specific-password  # NOT your regular Gmail password
SMTP_FROM=noreply@arc-consent.com

# How to get Gmail App Password:
# 1. Go to Google Account Settings
# 2. Security → 2-Step Verification
# 3. App passwords → Generate for "Mail"
# 4. Copy 16-character password
```

**Alternative SMTP Providers**:

```bash
# SendGrid
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASS=your-sendgrid-api-key

# Mailtrap (testing only)
SMTP_HOST=smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USER=your-mailtrap-user
SMTP_PASS=your-mailtrap-pass

# AWS SES
SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SMTP_PORT=587
SMTP_USER=your-ses-smtp-username
SMTP_PASS=your-ses-smtp-password
```

### Apply SMTP Configuration

**For Docker/K8s**:
```bash
# Update secret
kubectl create secret generic arc-backend-smtp \
  --from-literal=SMTP_USER='your-email@gmail.com' \
  --from-literal=SMTP_PASS='your-app-password' \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart pods to pick up new config
kubectl rollout restart deployment/arc-backend
```

**For Direct Deployment**:
```bash
# Edit .env file
nano /path/to/backend/.env.production

# Restart service
sudo systemctl restart arc-backend
```

---

## Testing After SMTP Configuration

```powershell
# Re-run test suite
.\test-backend-production.ps1
```

**Expected Results with SMTP Configured**:
- ✅ Test 1: Status 201, message "Account created successfully"
- ✅ Test 2: Status 403, "Email not verified"
- ✅ Test 3: Status 400, "User already exists"
- ✅ Test 4: Status 401, "Invalid credentials"
- ✅ Test 5: Status 201, guardian email sent

---

## Rollback Plan

If issues occur after deployment:

### Docker/K8s Rollback
```bash
# Rollback to previous deployment
kubectl rollout undo deployment/arc-backend

# Or specify revision
kubectl rollout undo deployment/arc-backend --to-revision=<previous-revision>
```

### Direct Deployment Rollback
```bash
# Restore previous binary
cp bin/server.backup bin/server
sudo systemctl restart arc-backend
```

---

## Monitoring

### Check Logs for License Errors
```bash
# Kubernetes
kubectl logs -f deployment/arc-backend | grep -i "license"

# Docker
docker logs -f arc-backend | grep -i "license"

# Direct
tail -f /var/log/arc-backend/app.log | grep -i "license"
```

**Expected**: No more "License required" errors for `/api/v1/auth/*` endpoints

### Check Logs for Email Errors
```bash
kubectl logs -f deployment/arc-backend | grep -i "email"
```

**Watch For**:
- "Failed to send verification email" → SMTP config issue
- "Email service unreachable" → SMTP connection issue
- "Email service authentication failed" → SMTP credentials wrong

---

## Success Criteria

- [ ] Health endpoint returns 200
- [ ] Auth signup endpoints return 503 (SMTP) or 201 (configured) - NOT 402
- [ ] Auth login endpoints return 401 (invalid) or 403 (unverified) - NOT 402
- [ ] SSO endpoints accessible without license error
- [ ] Email service sends verification emails (check spam folder)
- [ ] No license errors in logs for auth endpoints

---

## Support & Troubleshooting

### Issue: Still getting 402 errors

**Solution**: Backend not redeployed with new code
```bash
# Verify deployment updated
kubectl get pods -l app=arc-backend
# Check pod age - should be recent

# Force restart
kubectl rollout restart deployment/arc-backend
```

### Issue: 503 Email Service Unavailable

**Solution**: SMTP not configured
```bash
# Check environment variables loaded
kubectl exec -it <pod-name> -- env | grep SMTP

# Verify SMTP credentials are correct
# Test SMTP connection manually
```

### Issue: Emails not being received

**Solutions**:
1. Check spam/junk folder
2. Verify SMTP_FROM domain has SPF/DKIM records
3. Use Gmail App Password (not regular password)
4. Check email service logs for errors

---

## Contact

For deployment issues, check:
1. This deployment guide
2. `walkthrough.md` - Implementation details
3. `API_TESTING.md` - Test scenarios
4. Production logs
