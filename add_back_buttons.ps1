# Add back buttons to all ARES HTML pages
$pages = @(
    "solace-trading.html",
    "forge-dashboard.html", 
    "code-ide.html",
    "solace-control.html",
    "memory.html",
    "vision.html",
    "health.html"
)

$backButtonCSS = @"
        .back-btn {
            position: absolute;
            top: 20px;
            left: 30px;
            padding: 8px 16px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            display: flex;
            align-items: center;
            gap: 5px;
            font-size: 14px;
            transition: all 0.3s;
            text-decoration: none;
            z-index: 1000;
        }

        .back-btn:hover {
            background: #764ba2;
            transform: translateX(-3px);
        }

        .back-btn::before {
            content: '←';
            font-size: 18px;
        }
"@

$backButtonHTML = '<a href="/dashboard.html" class="back-btn">Back</a>'

foreach ($page in $pages) {
    $filePath = "C:\ARES_Workspace\ARES_API\web\$page"
    if (Test-Path $filePath) {
        Write-Host "Processing $page..." -ForegroundColor Cyan
        $content = Get-Content $filePath -Raw
        
        # Add CSS if not present
        if ($content -notmatch ".back-btn") {
            $content = $content -replace '(</style>)', "$backButtonCSS`n`$1"
            Write-Host "  ✅ Added back button CSS" -ForegroundColor Green
        }
        
        # Add button HTML if not present (look for common header patterns)
        if ($content -notmatch 'class="back-btn"') {
            # Try to find header div and add button
            $content = $content -replace '(<div class="header">)', "`$1`n                $backButtonHTML"
            Write-Host "  ✅ Added back button HTML" -ForegroundColor Green
        }
        
        Set-Content $filePath -Value $content -NoNewline
        Write-Host "  Done!" -ForegroundColor Green
    } else {
        Write-Host "❌ File not found: $page" -ForegroundColor Red
    }
}

Write-Host "`n🎉 All back buttons added!" -ForegroundColor Yellow
