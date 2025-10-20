# ARES Sensitive File Encryption

This repository uses encrypted sensitive configuration files to protect API keys, database credentials, and other secrets while allowing safe version control.

## Encrypted Files

- `.env.enc` - Main environment configuration (API keys, database credentials)
- `.env.claude.enc` - Claude API configuration

## Usage

### Decrypt Files (for development)

```powershell
# From ARES_API directory
..\encrypt_files.ps1 -Action decrypt
```

This will create:
- `.env` (from `.env.enc`)
- `.env.claude` (from `.env.claude.enc`)

### Encrypt Files (before commit)

```powershell
# From ARES_API directory
..\encrypt_files.ps1 -Action encrypt
```

This will:
- Encrypt `.env` → `.env.enc`
- Encrypt `.env.claude` → `.env.claude.enc`
- Create backups: `.env.backup`, `.env.claude.backup`

### Clean Up

After decryption, remember to delete the decrypted files before committing:

```powershell
Remove-Item .env, .env.claude
```

## Security Notes

- **Never commit** the decrypted `.env` or `.env.claude` files
- **Always encrypt** sensitive files before pushing to repository
- **Use strong passwords** if customizing the encryption key
- **Backup encrypted files** regularly

## What's Encrypted

The encrypted files contain:
- OpenAI API keys
- Anthropic Claude API keys
- Database passwords
- JWT secrets
- SOLACE API keys

## Emergency Decryption

If you lose the encryption script, you can manually decrypt using PowerShell:

```powershell
$encrypted = Get-Content ".env.enc" -Raw
$secureString = ConvertTo-SecureString $encrypted -Key (1..32 | ForEach-Object { [byte]$_ })
$BSTR = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secureString)
$decrypted = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($BSTR)
$decrypted | Out-File ".env" -Encoding UTF8
```