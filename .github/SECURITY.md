# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

The EngineerDNA team takes security vulnerabilities seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report a Security Vulnerability

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them using one of the following methods:

1. **GitHub Security Advisories (Preferred)**
   - Go to https://github.com/EngineerDNA/engineerdna/security/advisories/new
   - Click "Report a vulnerability"
   - Fill out the form with as much detail as possible

2. **Email**
   - Send an email to security@engineerdna.io
   - Include detailed information about the vulnerability
   - Use PGP encryption if possible (key available on request)

### What to Include in Your Report

Please include the following information in your vulnerability report:

- Type of vulnerability (e.g., SQL injection, XSS, authentication bypass)
- Full paths of source file(s) related to the vulnerability
- The location of the affected source code (tag/branch/commit or direct URL)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the vulnerability (what an attacker could achieve)
- Suggested fix or mitigation (if you have one)

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours
- **Updates**: We will provide regular updates on our progress at least every 7 days
- **Timeline**: We aim to:
  - Validate the vulnerability within 5 business days
  - Develop a fix within 30 days for critical vulnerabilities
  - Release a patched version as soon as safely possible
- **Credit**: We will credit you in the security advisory (unless you prefer to remain anonymous)

### Security Vulnerability Response Process

1. **Triage** (1-2 days): Verify and assess the severity
2. **Investigation** (3-5 days): Understand the root cause and impact
3. **Fix Development** (varies by severity): Develop and test the fix
4. **Coordinated Disclosure**: Work with you on disclosure timeline
5. **Release**: Publish the fix and security advisory
6. **Public Disclosure**: After the fix is released and users have had time to update

## Security Best Practices for EngineerDNA

### For Users

1. **Keep Updated**: Always run the latest version of EngineerDNA
2. **Localhost Only**: EngineerDNA v1 is designed for localhost use only (127.0.0.1)
3. **Master Key Security**: Store your master encryption key securely
   - Use OS keychain when possible
   - Never commit `.env` files with master keys
4. **Plugin Security**: Only use plugins from trusted sources
5. **Sensitive Data**: Review anonymization settings before exporting data

### For Plugin Developers

1. **Input Validation**: Always validate and sanitize user inputs
2. **Secrets Management**: Never log or expose API keys or tokens
3. **Subprocess Security**: Use timeouts and validate all subprocess inputs
4. **Network Security**: Use HTTPS for all external API calls
5. **Follow the Plugin SDK**: Use the official EngineerDNA Plugin SDK patterns

### For Contributors

1. **No Hardcoded Secrets**: Never commit API keys, tokens, or passwords
2. **Encryption Required**: All secrets must use AES-256-GCM encryption
3. **Localhost Binding**: Always bind to 127.0.0.1, never 0.0.0.0 (in v1)
4. **SQL Injection Prevention**: Use parameterized queries, never string concatenation
5. **Dependency Security**: Run `govulncheck` before submitting PRs
6. **Code Review**: Security-sensitive changes require security team review

## Security Features in EngineerDNA

### Current Security Measures (v1.x)

- **Network Isolation**: Localhost-only binding (127.0.0.1:3847)
- **Encryption at Rest**: AES-256-GCM for all secrets and API keys
- **Plugin Isolation**: Subprocess model with timeouts
- **Anonymization**: PII protection for external API transmission
- **Audit Logging**: All exports and processing operations logged
- **Input Validation**: Strict validation on all user inputs
- **SQL Injection Protection**: Parameterized queries only
- **Dependency Scanning**: Automated vulnerability checks in CI

### Planned Security Features (Future Versions)

- JWT authentication (when network access is added)
- Rate limiting
- API key management
- TLS/HTTPS support
- Role-based access control (RBAC)
- OAuth2 integration

## Known Security Limitations

### Version 1.x (Current)

1. **No Network Authentication**: Designed for single-user localhost use
   - Anyone with access to your machine can access the application
   - Mitigated by: OS-level authentication (login required)

2. **Plugin Trust Model**: Users must trust plugins they install
   - Plugins run with same privileges as main application
   - Mitigated by: Plugin isolation via subprocess model

3. **Anonymization Reversibility**: Anonymization mappings are stored locally
   - Allows deanonymization of exported data
   - Mitigated by: Audit logging and user awareness

## Security Updates

Security updates will be released as:
- **Critical**: Immediate patch release (within 24-48 hours)
- **High**: Patch release within 7 days
- **Medium**: Next minor release
- **Low**: Next minor or major release

Security advisories will be published at:
- GitHub Security Advisories: https://github.com/EngineerDNA/engineerdna/security/advisories
- Release notes: https://github.com/EngineerDNA/engineerdna/releases

## Bug Bounty Program

We do not currently have a bug bounty program, but we greatly appreciate responsible disclosure and will publicly acknowledge security researchers who help improve EngineerDNA's security.

## Questions?

If you have questions about this security policy, please open a discussion at:
https://github.com/EngineerDNA/engineerdna/discussions

For security-related questions, email: security@engineerdna.io

## Hall of Fame

We thank the following security researchers for their responsible disclosure:

(No vulnerabilities reported yet)
