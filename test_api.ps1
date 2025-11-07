# Script de test pour l'API URL Shortener

Write-Host "`n=== Test 1: Health Check ===" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Content: $($response.Content)" -ForegroundColor Green
} catch {
    Write-Host "Erreur: $_" -ForegroundColor Red
}

Write-Host "`n=== Test 2: Créer un lien court ===" -ForegroundColor Cyan
try {
    $body = @{
        long_url = "https://www.google.com"
    } | ConvertTo-Json

    $response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/links" `
        -Method POST `
        -Body $body `
        -ContentType "application/json" `
        -UseBasicParsing
    
    Write-Host "Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Content: $($response.Content)" -ForegroundColor Green
    
    # Extraire le short_code pour les tests suivants
    $result = $response.Content | ConvertFrom-Json
    $shortCode = $result.short_code
    
    Write-Host "`n=== Test 3: Obtenir les stats ===" -ForegroundColor Cyan
    $statsResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/links/$shortCode/stats" -UseBasicParsing
    Write-Host "Status: $($statsResponse.StatusCode)" -ForegroundColor Green
    Write-Host "Content: $($statsResponse.Content)" -ForegroundColor Green
    
    Write-Host "`n=== Test 4: Redirection ===" -ForegroundColor Cyan
    Write-Host "Vous pouvez tester la redirection en ouvrant: http://localhost:8080/$shortCode" -ForegroundColor Yellow
    
} catch {
    Write-Host "Erreur: $_" -ForegroundColor Red
}
