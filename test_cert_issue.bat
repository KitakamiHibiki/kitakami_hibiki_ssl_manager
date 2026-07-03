@echo off
setlocal enabledelayedexpansion

set DOMAIN=%1
set EMAIL=%2
if "%DOMAIN%"=="" set DOMAIN=test.kitakamihibiki.top
if "%EMAIL%"=="" set EMAIL=test@test.com

set OUT_DIR=test_certs_%date:~0,4%%date:~5,2%%date:~8,2%_%time:~0,2%%time:~3,2%%time:~6,2%
set OUT_DIR=%OUT_DIR: =0%

set "DNS_SERVERS=119.29.29.29:53,223.5.5.5:53,114.114.114.114:53,8.8.8.8:53"

echo [INFO] Checking lego...

where lego >nul 2>&1
if %errorlevel% neq 0 (
    echo [WARN] lego not found, installing via go install...
    set GOPROXY=https://goproxy.cn,direct
    go install github.com/go-acme/lego/v4/cmd/lego@latest
    if %errorlevel% neq 0 (
        echo [ERR] Failed to install lego
        exit /b 1
    )
    for /f "tokens=*" %%i in ('go env GOPATH') do set "PATH=!PATH!;%%i\bin"
)
echo [OK] lego ready

mkdir "%OUT_DIR%" 2>nul

echo [INFO] Requesting cert from Let's Encrypt: %DOMAIN%
lego --email "%EMAIL%" --domains "%DOMAIN%" --dns manual --dns.resolvers "%DNS_SERVERS%" --path "%OUT_DIR%\.lego" --accept-tos run > "%OUT_DIR%\lego_output.txt" 2>&1

powershell -Command "$txt=Get-Content '%OUT_DIR%\lego_output.txt' -Raw; if($txt -match '_acme-challenge\.(\S+)'){Write-Output $matches[1]}else{Write-Output 'NOTFOUND'}" > "%OUT_DIR%\tmp_fqdn.txt"
set /p CHALLENGE_FQDN=<"%OUT_DIR%\tmp_fqdn.txt"

powershell -Command "$txt=Get-Content '%OUT_DIR%\lego_output.txt' -Raw; if($txt -match 'IN TXT \""([^\""]+)\""'){Write-Output $matches[1]}else{Write-Output 'NOTFOUND'}" > "%OUT_DIR%\tmp_val.txt"
set /p CHALLENGE_VALUE=<"%OUT_DIR%\tmp_val.txt"

if "%CHALLENGE_FQDN%"=="NOTFOUND" (
    type "%OUT_DIR%\lego_output.txt"
    echo [ERR] Failed to get DNS challenge value
    exit /b 1
)

echo [OK] DNS challenge value obtained
echo.
echo ========================================
echo   FQDN : _acme-challenge.%CHALLENGE_FQDN%
echo   Value: %CHALLENGE_VALUE%
echo ========================================
echo.

set /p DUMMY=Add the above TXT record to DNS, then press Enter...

echo [INFO] Verifying DNS TXT record...
set VERIFIED=0
set DNS_SEC=0
for /l %%i in (1,1,60) do (
    set DNS_SEC=%%i
    powershell -Command "$fqdn='_acme-challenge.%CHALLENGE_FQDN%'; $servers=@('119.29.29.29','223.5.5.5','114.114.114.114','8.8.8.8'); foreach($s in $servers){$r=Resolve-DnsName -Name $fqdn -Type TXT -Server $s -ErrorAction SilentlyContinue; if($r){$r.Strings; break}}" > "%OUT_DIR%\dns_result.txt" 2>nul
    findstr /c:"%CHALLENGE_VALUE%" "%OUT_DIR%\dns_result.txt" >nul 2>&1
    if !errorlevel! equ 0 (
        set VERIFIED=1
        goto :dns_ok
    )
    echo   Waiting for DNS propagation... %%is / 60s
    timeout /t 1 >nul
)

:dns_ok
if %VERIFIED% equ 1 (
    echo [OK] DNS verified ^(%DNS_SEC%s^)
) else (
    echo [ERR] DNS TXT record not found ^(timeout 60s^)
    exit /b 1
)

echo [INFO] Submitting DNS challenge to Let's Encrypt...
lego --email "%EMAIL%" --domains "%DOMAIN%" --dns manual --dns.resolvers "%DNS_SERVERS%" --path "%OUT_DIR%\.lego" --accept-tos run

set CERT_DIR=%OUT_DIR%\.lego\certificates
if exist "%CERT_DIR%\%DOMAIN%.crt" (
    echo [OK] Certificate issued successfully!
    echo.
    echo ========================================
    echo   Cert: %CERT_DIR%\%DOMAIN%.crt
    echo   Key : %CERT_DIR%\%DOMAIN%.key
    echo ========================================
    copy "%CERT_DIR%\%DOMAIN%.crt" "%OUT_DIR%\%DOMAIN%.fullchain.pem" >nul
    type "%CERT_DIR%\%DOMAIN%.issuer.crt" >> "%OUT_DIR%\%DOMAIN%.fullchain.pem" 2>nul
    copy "%CERT_DIR%\%DOMAIN%.key" "%OUT_DIR%\%DOMAIN%.privkey.pem" >nul
    echo [OK] Copied to: %OUT_DIR%\
) else (
    echo [ERR] Certificate issuance failed
    exit /b 1
)

endlocal
