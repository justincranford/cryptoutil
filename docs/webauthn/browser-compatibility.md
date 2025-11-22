# WebAuthn Browser and Platform Compatibility

## Overview

This document provides comprehensive browser and platform compatibility information for WebAuthn (Web Authentication API) used in the cryptoutil identity platform. WebAuthn enables passwordless authentication using platform authenticators (Windows Hello, TouchID, FaceID, Android Biometric) and external authenticators (FIDO2 security keys).

## Browser Support Matrix

### Desktop Browsers

| Browser | Minimum Version | Platform Authenticator | External Authenticator | Notes |
|---------|----------------|------------------------|------------------------|-------|
| **Chrome** | 67+ (June 2018) | ✅ Yes | ✅ Yes | Full WebAuthn Level 2 support since v90 |
| **Edge** | 18+ (October 2018) | ✅ Yes | ✅ Yes | Chromium-based Edge (79+) recommended |
| **Firefox** | 60+ (May 2018) | ✅ Yes | ✅ Yes | Full support since v77 |
| **Safari** | 13+ (September 2019) | ✅ Yes | ✅ Yes | macOS Catalina 10.15+ required for TouchID |
| **Opera** | 54+ (September 2018) | ✅ Yes | ✅ Yes | Chromium-based, follows Chrome support |

### Mobile Browsers

| Browser | Platform | Minimum Version | Platform Authenticator | External Authenticator | Notes |
|---------|----------|----------------|------------------------|------------------------|-------|
| **Chrome** | Android | 70+ (October 2018) | ✅ Yes | ⚠️ Limited | Android 7+ with biometric hardware |
| **Chrome** | iOS | 89+ (March 2021) | ❌ No | ⚠️ Limited | iOS WebAuthn support limited |
| **Safari** | iOS | 14+ (September 2020) | ✅ Yes | ⚠️ Limited | FaceID/TouchID on supported devices |
| **Edge** | Android | 45+ (April 2020) | ✅ Yes | ⚠️ Limited | Android 7+ with biometric hardware |
| **Samsung Internet** | Android | 13+ (January 2020) | ✅ Yes | ⚠️ Limited | Android 8+ with biometric hardware |

### Browser Feature Support

| Feature | Chrome | Edge | Firefox | Safari | Status |
|---------|--------|------|---------|--------|--------|
| WebAuthn Level 1 | ✅ v67+ | ✅ v18+ | ✅ v60+ | ✅ v13+ | Stable |
| WebAuthn Level 2 | ✅ v90+ | ✅ v90+ | ✅ v77+ | ✅ v14+ | Stable |
| Resident Keys (Passkeys) | ✅ v90+ | ✅ v90+ | ✅ v77+ | ✅ v14+ | Stable |
| User Verification | ✅ v67+ | ✅ v18+ | ✅ v60+ | ✅ v13+ | Stable |
| Attestation (direct) | ✅ v67+ | ✅ v18+ | ✅ v60+ | ⚠️ Limited | Safari privacy restrictions |
| Cross-Origin Support | ✅ v90+ | ✅ v90+ | ⚠️ Limited | ⚠️ Limited | Chrome/Edge best support |

## Platform Authenticator Support

### Windows (Windows Hello)

**Requirements:**
- Windows 10 version 1903+ (May 2019 Update) or Windows 11
- TPM 2.0 chip (hardware security module)
- PIN or biometric setup (fingerprint, facial recognition, iris)

**Browser Compatibility:**
- ✅ Chrome 67+
- ✅ Edge 18+ (Chromium-based Edge 79+ recommended)
- ✅ Firefox 60+
- ⚠️ Opera 54+ (limited testing)

**Configuration:**
- Settings → Accounts → Sign-in options → Windows Hello
- Requires device enrollment with Microsoft Account or Azure AD
- Supports both biometric and PIN-based authentication

### macOS (TouchID)

**Requirements:**
- macOS Catalina 10.15+ (September 2019)
- Mac with TouchID sensor (MacBook Pro 2016+, MacBook Air 2018+, Mac Mini 2018+)
- iCloud Keychain enabled (for credential sync)

**Browser Compatibility:**
- ✅ Safari 13+
- ✅ Chrome 70+
- ✅ Edge 79+
- ⚠️ Firefox 60+ (requires manual enablement via about:config)

**Configuration:**
- System Preferences → Touch ID → Add fingerprint
- Safari automatically uses TouchID for WebAuthn
- Chrome/Edge require iCloud Keychain integration

### iOS (FaceID/TouchID)

**Requirements:**
- iOS 14.0+ (September 2020) for WebAuthn support
- iPhone/iPad with FaceID (iPhone X+, iPad Pro 2018+) or TouchID (iPhone 5s+, iPad Air 2+)
- iCloud Keychain enabled for credential sync

**Browser Compatibility:**
- ✅ Safari 14+
- ⚠️ Chrome 89+ (limited support, uses Safari WebKit)
- ⚠️ Edge 45+ (limited support, uses Safari WebKit)

**Limitations:**
- Third-party browsers use Safari WebKit engine (Apple requirement)
- WebAuthn support limited compared to Safari
- Cross-origin WebAuthn not fully supported

### Android (Biometric Authenticator)

**Requirements:**
- Android 7.0+ (Nougat, August 2016) for basic WebAuthn
- Android 9.0+ (Pie, August 2018) for full FIDO2 support
- Device with biometric hardware (fingerprint, face unlock, iris)
- Screen lock configured (PIN, pattern, password, or biometric)

**Browser Compatibility:**
- ✅ Chrome 70+ (recommended)
- ✅ Edge 45+
- ⚠️ Samsung Internet 13+
- ❌ Firefox Android (no platform authenticator support as of v100)

**Configuration:**
- Settings → Security → Biometric & security → Fingerprints/Face recognition
- Chrome uses Android Keystore for credential storage
- Supports FIDO2 security keys via USB/NFC/Bluetooth

## External Authenticator Support

### FIDO2 Security Keys

**Supported Protocols:**
- USB HID (Human Interface Device)
- NFC (Near Field Communication)
- Bluetooth Low Energy (BLE)

**Compatible Devices:**
- YubiKey 5 Series (USB-A, USB-C, NFC, Lightning)
- Google Titan Security Key (USB-A, USB-C, NFC, Bluetooth)
- Feitian ePass FIDO2 Security Keys
- Solo Keys (open-source FIDO2)
- Windows Hello for Business (USB-based)

**Browser Compatibility:**
- ✅ Chrome 67+ (all transport types)
- ✅ Edge 18+ (all transport types)
- ✅ Firefox 60+ (USB, NFC; BLE requires manual enablement)
- ✅ Safari 13+ (USB, NFC; BLE limited)

**Mobile Compatibility:**
- ✅ Android Chrome 70+ (USB OTG, NFC, Bluetooth)
- ⚠️ iOS Safari 14+ (NFC only, requires iPhone 7+)

## Fallback Strategies

### Strategy 1: Feature Detection with Graceful Degradation

```javascript
if (window.PublicKeyCredential && 
    PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable) {
    // WebAuthn supported, check for platform authenticator
    PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable()
        .then(available => {
            if (available) {
                // Offer platform authenticator (Windows Hello, TouchID, etc.)
                enableWebAuthnRegistration();
            } else {
                // Offer external authenticator (FIDO2 security key)
                enableSecurityKeyRegistration();
            }
        });
} else {
    // WebAuthn not supported, fall back to traditional authentication
    enablePasswordAuthentication();
}
```

### Strategy 2: Progressive Enhancement

**Tier 1: WebAuthn + Platform Authenticator (Best Experience)**
- Windows 10/11 with Windows Hello
- macOS 10.15+ with TouchID
- iOS 14+ with FaceID/TouchID
- Android 7+ with biometric hardware

**Tier 2: WebAuthn + External Authenticator (Security Keys)**
- Desktop browsers with USB/NFC security keys
- Mobile browsers with NFC security keys (Android/iOS)

**Tier 3: Traditional Multi-Factor Authentication**
- TOTP (Time-based One-Time Password)
- SMS OTP (fallback for unsupported browsers)
- Email verification links

**Tier 4: Password-Only (Legacy Fallback)**
- Browsers without WebAuthn support (<2% market share as of 2024)
- Enterprise environments with restrictive browser policies

### Strategy 3: Browser-Specific Recommendations

**For Safari iOS Users:**
- Recommend Safari browser (best WebAuthn support)
- Warn about limited third-party browser support
- Suggest updating to iOS 14+ for WebAuthn availability

**For Firefox Android Users:**
- Recommend Chrome browser (platform authenticator support)
- Offer FIDO2 security key as alternative
- Provide TOTP/SMS fallback

**For Enterprise Users (IE11, Old Edge):**
- Detect legacy browsers with User-Agent sniffing
- Display upgrade recommendation banner
- Provide password + TOTP fallback

### Strategy 4: Error Handling and User Guidance

**Common WebAuthn Errors:**

| Error | Cause | User Guidance |
|-------|-------|---------------|
| `NotSupportedError` | Browser lacks WebAuthn support | "Please use a modern browser (Chrome, Edge, Firefox, Safari)" |
| `NotAllowedError` | User cancelled ceremony | "Authentication cancelled. Please try again." |
| `InvalidStateError` | Credential already registered | "This authenticator is already registered. Use it to sign in." |
| `SecurityError` | HTTPS required or origin mismatch | "WebAuthn requires a secure connection (HTTPS)" |
| `AbortError` | Timeout or user inactivity | "Authentication timed out. Please try again." |
| `UnknownError` | Authenticator malfunction | "Authenticator error. Try a different device or method." |

**Recommended User Flow:**
1. Detect WebAuthn availability before showing registration option
2. Provide clear instructions: "Touch your fingerprint sensor" or "Insert your security key"
3. Show fallback options immediately if WebAuthn fails
4. Log errors for monitoring and troubleshooting

## Testing and Validation

### Desktop Browser Testing

**Chrome/Edge:**
```bash
# Enable virtual authenticator for testing
chrome://flags/#enable-web-authentication-testing-api
```

**Firefox:**
```bash
# Enable FIDO2 debugging
about:config → security.webauthn.ctap2 = true
about:config → security.webauthn.enable_uv_level2 = true
```

**Safari:**
```bash
# Enable Develop menu
Safari → Preferences → Advanced → Show Develop menu
# Inspect WebAuthn API via console
```

### Mobile Browser Testing

**iOS Safari:**
- Requires physical device with FaceID/TouchID (simulator lacks biometric hardware)
- Test via Safari Technology Preview for early feature access
- Use Xcode device logs for debugging

**Android Chrome:**
- Emulator supports virtual biometric authentication
- Settings → Security → Fingerprint → Enable emulated fingerprint
- Use Chrome DevTools remote debugging

### Automated Testing

**WebAuthn Virtual Authenticator API (Chrome/Edge):**
```javascript
// Create virtual authenticator for E2E testing
const authenticator = await driver.executeScript(`
  return window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable();
`);

// Selenium WebDriver 4+ with virtual authenticator support
driver.addVirtualAuthenticator(new VirtualAuthenticatorOptions()
  .setProtocol(VirtualAuthenticatorOptions.Protocol.CTAP2)
  .setTransport(VirtualAuthenticatorOptions.Transport.INTERNAL)
  .setHasUserVerification(true));
```

## Browser Update Recommendations

**Critical Security Updates:**
- Chrome: Update to latest stable (automatic updates enabled by default)
- Edge: Update to Chromium-based Edge 79+ (replaces legacy Edge)
- Firefox: Update to v77+ for WebAuthn Level 2 support
- Safari: Update to macOS 10.15+ and Safari 13+ for TouchID

**Enterprise Deployment:**
- Minimum versions: Chrome 90+, Edge 90+, Firefox 77+, Safari 14+
- Block WebAuthn enrollment on unsupported browsers
- Provide TOTP/SMS fallback for legacy browser users

## Privacy Considerations

### Attestation Restrictions

**Safari Privacy Mode:**
- Safari restricts direct attestation (privacy protection)
- Only anonymous attestation allowed by default
- Relying Party cannot determine authenticator make/model

**Firefox Enhanced Tracking Protection:**
- May block cross-origin WebAuthn requests
- Requires Same-Site cookie policy for credential storage
- User must whitelist site for persistent credentials

**Chrome Incognito/Private Browsing:**
- Credentials not saved across sessions
- Platform authenticator works (Windows Hello, TouchID)
- External authenticators require re-registration per session

## Migration and Upgrade Paths

### From Password-Only to WebAuthn

**Phase 1: Opt-In WebAuthn (Pilot)**
- Offer WebAuthn as optional authentication method
- Keep password authentication available
- Target early adopters with modern browsers

**Phase 2: Recommended WebAuthn (Rollout)**
- Prompt users to register WebAuthn after password login
- Provide browser compatibility checker
- Maintain password + TOTP as fallback

**Phase 3: WebAuthn-First (Passwordless)**
- Default to WebAuthn for new user registrations
- Allow password reset for WebAuthn-enabled users
- Deprecate password-only accounts gradually

### From Legacy MFA to WebAuthn

**Migration Steps:**
1. Enroll WebAuthn authenticator (platform or security key)
2. Verify WebAuthn credential works correctly
3. Optionally disable TOTP/SMS after successful WebAuthn use
4. Maintain at least 2 WebAuthn credentials (backup)

**Rollback Plan:**
- Keep existing TOTP/SMS methods active during transition
- Allow re-enabling password authentication if needed
- Provide support for lost/damaged authenticators

## Compliance and Standards

### FIDO2 Certification

**Certified Authenticators:**
- YubiKey 5 Series (FIDO2 Level 1 and Level 2)
- Windows Hello (FIDO2 Certified)
- Google Titan Security Keys (FIDO2 Level 1)
- Apple TouchID/FaceID (FIDO2 compliant, not certified)

**Relying Party Certification:**
- cryptoutil follows FIDO2 Server Requirements
- Implements WebAuthn Level 2 specification (W3C Recommendation)
- Supports both user verification and user presence

### Regulatory Compliance

**GDPR (EU General Data Protection Regulation):**
- WebAuthn credentials stored locally on device (privacy-preserving)
- Server stores only public keys (no biometric data)
- User consent required before credential registration

**PSD2 (Payment Services Directive 2):**
- WebAuthn qualifies as Strong Customer Authentication (SCA)
- Meets "inherence" factor (biometrics) or "possession" factor (security key)
- Combined with knowledge factor (PIN) for two-factor authentication

**NIST 800-63B (Digital Identity Guidelines):**
- WebAuthn with platform authenticator: AAL3 (Authenticator Assurance Level 3)
- WebAuthn with security key: AAL3 with hardware-backed key storage
- Replay attack prevention via sign counter validation

## Troubleshooting Common Issues

### Issue 1: "WebAuthn not supported" on Supported Browser

**Symptoms:**
- `window.PublicKeyCredential` is undefined
- Feature detection returns false

**Causes:**
- HTTP connection (HTTPS required for WebAuthn)
- Browser extension blocking WebAuthn API
- Enterprise group policy disabling WebAuthn

**Solutions:**
- Ensure HTTPS connection with valid TLS certificate
- Disable browser extensions (test in incognito mode)
- Check enterprise policies: `chrome://policy` or `about:policies`

### Issue 2: Platform Authenticator Not Available

**Symptoms:**
- `isUserVerifyingPlatformAuthenticatorAvailable()` returns false
- "No authenticator found" error

**Causes:**
- Windows Hello/TouchID not configured
- TPM 2.0 chip missing (Windows)
- macOS version too old (pre-Catalina)

**Solutions:**
- Windows: Settings → Accounts → Sign-in options → Set up Windows Hello
- macOS: System Preferences → Touch ID → Add fingerprint
- Upgrade OS to supported version (Windows 10 1903+, macOS 10.15+)

### Issue 3: Security Key Not Detected

**Symptoms:**
- User inserted security key, but browser shows "waiting for authenticator"
- Timeout error after 60 seconds

**Causes:**
- USB port not providing sufficient power
- NFC not enabled on mobile device
- Security key firmware outdated

**Solutions:**
- Try different USB port (avoid USB hub)
- Enable NFC: Android Settings → Connected devices → Connection preferences → NFC
- Update security key firmware (YubiKey Manager, Feitian tools)

### Issue 4: Credential Registration Fails on iOS Safari

**Symptoms:**
- Registration succeeds on desktop, fails on iOS Safari
- "NotAllowedError: The operation is not allowed" error

**Causes:**
- Cross-origin iframe blocking WebAuthn
- Safari privacy restrictions on third-party cookies
- iOS version too old (<14.0)

**Solutions:**
- Avoid using WebAuthn in iframes (use top-level navigation)
- Enable Same-Site cookie policy for credential storage
- Require iOS 14+ for WebAuthn enrollment

## Future Roadmap

### Emerging Standards (2024-2026)

**WebAuthn Level 3 (W3C Draft):**
- Enhanced device attestation formats
- Improved privacy-preserving techniques
- Better support for hybrid authenticators (phone as security key)

**Passkeys (FIDO Alliance):**
- Cloud-synced credentials across devices (Apple Keychain, Google Password Manager)
- QR code-based cross-device authentication
- Improved user experience for passwordless login

**Conditional UI:**
- Browser autofill integration for passkeys
- Seamless credential selection without separate UI
- Improved mobile browser support

### Platform Improvements

**Windows:**
- Windows 11 enhanced biometric security
- Azure AD integration for enterprise passkeys
- Improved USB security key management

**macOS:**
- iCloud Keychain passkey sync across Apple devices
- Universal 2FA support with security keys
- Enhanced Safari WebAuthn debugging tools

**Android:**
- Google Play Services FIDO2 API updates
- Improved biometric authentication reliability
- Better NFC security key support

**iOS:**
- Passkey sync via iCloud Keychain
- Third-party browser improvements (WebKit updates)
- Enhanced FaceID/TouchID WebAuthn integration

## References

- [W3C WebAuthn Specification](https://www.w3.org/TR/webauthn-2/)
- [FIDO Alliance Standards](https://fidoalliance.org/specifications/)
- [MDN Web Docs: Web Authentication API](https://developer.mozilla.org/en-US/docs/Web/API/Web_Authentication_API)
- [Can I Use: WebAuthn](https://caniuse.com/webauthn)
- [FIDO2 Project by Google](https://github.com/google/fido2-net-lib)
- [YubiKey Developer Documentation](https://developers.yubico.com/WebAuthn/)
- [Microsoft Windows Hello Documentation](https://docs.microsoft.com/en-us/windows/security/identity-protection/hello-for-business/)
