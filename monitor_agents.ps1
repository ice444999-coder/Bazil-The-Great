# Live Agent Swarm Monitor
# Watch agents collaborate in real-time

Write-Host "`n╔═══════════════════════════════════════════════════════════════════╗" -ForegroundColor Magenta
Write-Host "║" -NoNewline -ForegroundColor Magenta
Write-Host "          AGENT SWARM LIVE COLLABORATION MONITOR               " -NoNewline -ForegroundColor White
Write-Host "║" -ForegroundColor Magenta
Write-Host "╚═══════════════════════════════════════════════════════════════════╝" -ForegroundColor Magenta

$logFile = "C:\ARES_Workspace\ARES_API\agent_coordinator.log"
$lastSize = 0
$agentColors = @{
    "SOLACE" = "Blue"
    "ARCHITECT" = "Cyan"
    "FORGE" = "Yellow"
    "SENTINEL" = "Green"
}

Write-Host "`n🔍 Monitoring: $logFile" -ForegroundColor Gray
Write-Host "Press Ctrl+C to stop`n" -ForegroundColor Gray

while ($true) {
    if (Test-Path $logFile) {
        $currentSize = (Get-Item $logFile).Length
        
        if ($currentSize -gt $lastSize) {
            $content = Get-Content $logFile -Encoding UTF8
            $newLines = $content[($lastSize/100)..[math]::Max(0, $content.Count-1)]
            
            foreach ($line in $newLines) {
                $color = "White"
                $prefix = ""
                
                if ($line -match "SOLACE") {
                    $color = $agentColors["SOLACE"]
                    $prefix = "🧠 "
                }
                elseif ($line -match "ARCHITECT") {
                    $color = $agentColors["ARCHITECT"]
                    $prefix = "📐 "
                }
                elseif ($line -match "FORGE") {
                    $color = $agentColors["FORGE"]
                    $prefix = "🔨 "
                }
                elseif ($line -match "SENTINEL") {
                    $color = $agentColors["SENTINEL"]
                    $prefix = "🛡️ "
                }
                elseif ($line -match "ERROR") {
                    $color = "Red"
                    $prefix = "❌ "
                }
                elseif ($line -match "completed") {
                    $color = "Green"
                    $prefix = "✓ "
                }
                elseif ($line -match "DELEGATE") {
                    $color = "Magenta"
                    $prefix = "➜ "
                }
                
                if ($line.Trim()) {
                    Write-Host "$prefix$line" -ForegroundColor $color
                }
            }
            
            $lastSize = $currentSize
        }
    }
    
    Start-Sleep -Seconds 2
}
