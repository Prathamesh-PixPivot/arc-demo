# Backend API Testing Plan - cURL Commands

## Prerequisites
- Backend running on `http://localhost:8080`
- SMTP configured (or testing with broken SMTP to verify rollback)
- Fresh database state

---

## Test 1: Manual Fiduciary Signup (Happy Path)

### Signup Request
```bash
curl -X POST http://localhost:8080/api/v1/auth/fiduciary/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test-fiduciary@example.com",
    "firstName": "Test",
    "lastName": "Fiduciary",
    "phone": "+1234567890",
    "password": "TestPass123!",
    "confirmPassword": "TestPass123!",
    "role": "admin",
    "organization": {
      "name": "Test Org",
      "industry": "technology",
      "companySize": "11-50",
      "address": "123 Test St",
      "country": "US",
      "email": "org@example.com",
      "phone": "+1234567890"
    }
  }'
```

**Expected Response** (200 Created):
```json
{
  "message": "Account created successfully. Please check your email for a verification link."
}
```

**Expected Behavior**:
- ✅ User created in database
- ✅ Organization created
- ✅ Tenant database created
- ✅ Verification email sent
- ❌ If email fails → entire signup rolled back

---

## Test 2: Manual Fiduciary Signup (SMTP Failure)

**Note**: Temporarily break SMTP config in `.env` to test rollback

### Broken SMTP Config
```env
SMTP_HOST=invalid-smtp-server.com
SMTP_PORT=587
SMTP_USER=invalid@example.com
SMTP_PASS=wrong-password
```

### Signup Request (Same as Test 1)

**Expected Response** (503 Service Unavailable):
```json
{
  "error": "Email service is temporarily unavailable. Please try again in a few minutes."
}
```

**Expected Behavior**:
- ❌ No user created
- ❌ No organization created
- ❌ No tenant database created
- ✅ Complete rollback

**Verification**:
```bash
# Check database - should NOT find user
curl http://localhost:8080/api/v1/auth/fiduciary/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test-fiduciary@example.com","password":"TestPass123!"}'
# Expected: 401 Unauthorized
```

---

## Test 3: Login Before Email Verification

### Restore SMTP config, create user (Test 1), then attempt login

### Login Request
```bash
curl -X POST http://localhost:8080/api/v1/auth/fiduciary/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test-fiduciary@example.com",
    "password": "TestPass123!"
  }'
```

**Expected Response** (403 Forbidden):
```json
{
  "error": "Email not verified. Please check your inbox for the verification link."
}
```

---

## Test 4: Email Verification

### Get verification token from email or database logs

### Verify Request
```bash
curl "http://localhost:8080/api/v1/auth/verify-fiduciary?token=<VERIFICATION_TOKEN>"
```

**Expected Response** (200 OK):
```json
{
  "message": "Account successfully verified."
}
```

---

## Test 5: Login After Verification

### Login Request (Same as Test 3)

**Expected Response** (200 OK):
```json
{
  "token": "eyJhbGc...",
  "fiduciaryId": "uuid-here",
  "email": "test-fiduciary@example.com",
  "phone": "+1234567890",
  "tenantId": "uuid-here",
  "expiresIn": 3600,
  "roles": ["Super Admin"],
  "permissions": { "manage_users": true, ... }
}
```

---

## Test 6: Manual User (Data Principal) Signup - Adult

### Signup Request
```bash
curl -X POST http://localhost:8080/api/v1/auth/user/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "adult-user@example.com",
    "password": "UserPass123!",
    "firstName": "Adult",
    "lastName": "User",
    "age": 25,
    "phone": "+9876543210",
    "location": "New York, USA"
  }'
```

**Expected Response** (200 Created):
```json
{
  "dataPrincipalId": "uuid-here",
  "message": "Account created successfully."
}
```

**Expected Behavior**:
- ✅ User created (IsVerified = true for adults)
- ❌ No guardian email sent

---

## Test 7: Manual User Signup - Minor (Guardian Email)

### Signup Request
```bash
curl -X POST http://localhost:8080/api/v1/auth/user/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "minor-user@example.com",
    "password": "MinorPass123!",
    "firstName": "Minor",
    "lastName": "User",
    "age": 15,
    "phone": "+1111111111",
    "guardianEmail": "guardian@example.com"
  }'
```

**Expected Response** (200 Created):
```json
{
  "dataPrincipalId": "uuid-here",
  "message": "Data principal created. If a minor, a verification email has been sent to the guardian."
}
```

**Expected Behavior**:
- ✅ User created (IsVerified = false)
- ✅ Guardian verification email sent
- ❌ If email fails → user creation rolled back

---

## Test 8: User Login Before Guardian Verification

### Login Request
```bash
curl -X POST http://localhost:8080/api/v1/auth/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "minor-user@example.com",
    "password": "MinorPass123!"
  }'
```

**Expected Response** (403 Forbidden):
```json
{
  "error": "account not verified. Please check your email."
}
```

---

## Test 9: SSO Login - Non-Existent User (STRICT MODE)

### Google SSO Login (Simulated)
```bash
# Step 1: Initiate SSO
curl "http://localhost:8080/api/v1/auth/sso/google?mode=login&userType=fiduciary"
# This will redirect to Google OAuth

# Step 2: After OAuth callback with non-existent email
# Expected: Redirect to frontend with error
# URL: http://localhost:3000/login?error=not_registered&hint=Please+sign+up+first
```

**Expected Behavior**:
- ❌ No user auto-created
- ✅ Redirects to login with error message
- ✅ User must explicitly use signup mode

---

## Test 10: SSO Signup - Fiduciary

### Google SSO Signup
```bash
curl "http://localhost:8080/api/v1/auth/sso/google?mode=signup&userType=fiduciary"
# After OAuth callback with email: sso-fiduciary@gmail.com
```

**Expected Behavior**:
- ✅ User created with `IsVerified = true`
- ✅ `AuthProvider = "google"`
- ✅ Redirect to `/onboarding/organization`

**Expected Redirect**:
```
http://localhost:3000/auth/callback?token=eyJhbGc...&userType=fiduciary&next=/onboarding/organization
```

---

## Test 11: Duplicate Signup Prevention

### Try to signup again with existing email

### Request (Same email as Test 1)
```bash
curl -X POST http://localhost:8080/api/v1/auth/fiduciary/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test-fiduciary@example.com",
    ...
  }'
```

**Expected Response** (400 Bad Request):
```json
{
  "error": "A user with this email already exists."
}
```

---

## Test 12: Invalid Credentials

### Login with wrong password

```bash
curl -X POST http://localhost:8080/api/v1/auth/fiduciary/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test-fiduciary@example.com",
    "password": "WrongPassword123!"
  }'
```

**Expected Response** (401 Unauthorized):
```json
{
  "error": "invalid credentials"
}
```

---

## Verification Checklist

| Test Case | Expected Result | Status |
|-----------|----------------|---------|
| Fiduciary Signup (Happy Path) | ✅ User created, email sent | ⏳ |
| Fiduciary Signup (SMTP Fail) | ✅ Complete rollback, no user | ⏳ |
| Login Before Verification | ❌ 403 Forbidden | ⏳ |
| Email Verification | ✅ Account verified | ⏳ |
| Login After Verification | ✅ Token returned | ⏳ |
| User Signup (Adult) | ✅ User created, no guardian email | ⏳ |
| User Signup (Minor) | ✅ Guardian email sent | ⏳ |
| Minor Login (Unverified) | ❌ 403 Forbidden | ⏳ |
| SSO Login (Non-Existent) | ❌ Redirect with error | ⏳ |
| SSO Signup (Fiduciary) | ✅ Redirect to onboarding | ⏳ |
| Duplicate Signup | ❌ 400 Bad Request | ⏳ |
| Invalid Credentials | ❌ 401 Unauthorized | ⏳ |

---

## Test Execution Order

1. **SMTP Configuration Check** → Verify email service is working
2. **Happy Path Tests** → Test 1, 4, 5, 6
3. **Error Handling Tests** → Test 2, 3, 7, 8, 11, 12
4. **SSO Tests** → Test 9, 10 (requires browser or modified testing)

---

## Notes

- Replace `<VERIFICATION_TOKEN>` with actual token from email/logs
- SSO tests require actual OAuth flow or mock server
- Check logs for email send confirmations
- Verify database state after each rollback test
