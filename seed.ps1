# Seed demo data for Lab Management System
param([int]$port = 8080)

$base = "http://localhost:$port/api/v1"
Write-Host "=== Logging in ==="
$loginBody = @{username="admin";password="admin123"} | ConvertTo-Json
$resp = Invoke-RestMethod -Uri "$base/auth/login" -Method Post -ContentType "application/json" -Body $loginBody
$token = $resp.data.token
$headers = @{Authorization="Bearer $token"; "Content-Type"="application/json"}

$users = @(
    @{username="zhangwei";   password="123456"; real_name="张伟"; email="zhangwei@lab.edu.cn";   phone="13800138001"; role_id=2},
    @{username="lina";       password="123456"; real_name="李娜"; email="lina@lab.edu.cn";       phone="13800138002"; role_id=3},
    @{username="wangqiang";  password="123456"; real_name="王强"; email="wangqiang@lab.edu.cn";  phone="13800138003"; role_id=3},
    @{username="zhaomin";    password="123456"; real_name="赵敏"; email="zhaomin@lab.edu.cn";    phone="13800138004"; role_id=3},
    @{username="chenjing";   password="123456"; real_name="陈静"; email="chenjing@lab.edu.cn";   phone="13800138005"; role_id=3}
)
Write-Host "Creating users..."
foreach ($u in $users) {
    $body = $u | ConvertTo-Json
    try {
        $r = Invoke-RestMethod -Uri "$base/users" -Method Post -Headers $headers -Body $body
        $ok = if ($r.code -eq 0) { "OK" } else { "SKIP" }
        Write-Host "  $ok $($u.username)"
    } catch { Write-Host "  ERR $($u.username)" }
}

$equip = @(
    @{name="示波器 TDS2024C";        model="TDS2024C";             category="测量仪器"; total_stock=3;  location="A301实验室"},
    @{name="GPU服务器 DGX-A100";     model="DGX-A100";             category="服务器";   total_stock=4;  location="A302机房"},
    @{name="3D打印机 Ultimaker S5";  model="Ultimaker S5";         category="制造设备"; total_stock=2;  location="B101创客空间"},
    @{name="频谱分析仪 N9020B";      model="Keysight N9020B";      category="测量仪器"; total_stock=1;  location="A303实验室"},
    @{name="MacBook Pro M3";         model="MacBook Pro 16 M3 Pro";category="笔记本电脑";total_stock=10; location="B202设备室"},
    @{name="逻辑分析仪 16862A";      model="Keysight 16862A";      category="测量仪器"; total_stock=2;  location="A304实验室"}
)
Write-Host "Creating equipment..."
foreach ($e in $equip) {
    $body = $e | ConvertTo-Json
    try {
        $r = Invoke-RestMethod -Uri "$base/equipments" -Method Post -Headers $headers -Body $body
        $ok = if ($r.code -eq 0) { "OK" } else { "SKIP" }
        Write-Host "  $ok $($e.name)"
    } catch { Write-Host "  ERR $($e.name)" }
}

Write-Host "Seed complete!"