# Скрипт для быстрого тестирования API
# Использование: .\test_api.ps1 <jwt_token>

param(
    [Parameter(Mandatory=$true)]
    [string]$Token
)

$baseUrl = "http://localhost:8080"
$userId = "550e8400-e29b-41d4-a716-446655440000"
$headers = @{
    "Authorization" = "Bearer $Token"
    "Content-Type" = "application/json"
}

Write-Host "=== Тестирование API ===" -ForegroundColor Green
Write-Host ""

# 1. Получить статус пользователя
Write-Host "1. GET /users/{id}/status" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/$userId/status" -Method Get -Headers $headers
    Write-Host "Успешно! Баланс: $($response.user.balance), Поинтов: $($response.total_points)" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 5
} catch {
    Write-Host "Ошибка: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 2. Получить лидерборд
Write-Host "2. GET /users/leaderboard" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/leaderboard?limit=5" -Method Get -Headers $headers
    Write-Host "Успешно! Найдено пользователей: $($response.leaderboard.Count)" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 3
} catch {
    Write-Host "Ошибка: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 3. Выполнить задание
Write-Host "3. POST /users/{id}/task/complete (telegram_subscribe)" -ForegroundColor Yellow
$taskBody = @{
    task_type = "telegram_subscribe"
} | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/$userId/task/complete" -Method Post -Headers $headers -Body $taskBody
    Write-Host "Успешно! Получено поинтов: $($response.points)" -ForegroundColor Green
    $response | ConvertTo-Json
} catch {
    Write-Host "Ошибка: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Детали: $responseBody" -ForegroundColor Red
    }
}
Write-Host ""

# 4. Установить реферера
Write-Host "4. POST /users/{id}/referrer" -ForegroundColor Yellow
$referrerBody = @{
    referrer_id = "550e8400-e29b-41d4-a716-446655440001"
} | ConvertTo-Json
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/users/$userId/referrer" -Method Post -Headers $headers -Body $referrerBody
    Write-Host "Успешно! $($response.message)" -ForegroundColor Green
    $response | ConvertTo-Json
} catch {
    Write-Host "Ошибка: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Детали: $responseBody" -ForegroundColor Red
    }
}
Write-Host ""

Write-Host "=== Тестирование завершено ===" -ForegroundColor Green

