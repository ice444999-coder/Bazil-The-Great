# Approach 1: PowerShell Function Wrapper
# Usage: .\db_query.ps1 "SELECT * FROM decision_traces LIMIT 1;"

param(
    [Parameter(Mandatory=$true)]
    [string]$Query
)

$env:PGPASSWORD = 'ARESISWAKING'
$result = & 'C:\Program Files\PostgreSQL\18\bin\psql.exe' -U ARES -d ares_db -t -A -c $Query
Write-Output $result
