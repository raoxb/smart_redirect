#!/bin/bash

echo "=== Smart Redirect 前端功能测试 ==="
echo ""

# 前端和后端URL
FRONTEND_URL="http://localhost:3001"
API_URL="http://localhost:3001/api/v1"

echo "1. 测试前端页面可访问性"
echo "   - 检查首页: $FRONTEND_URL"
if curl -s $FRONTEND_URL | grep -q "Smart Redirect"; then
    echo "   ✅ 前端首页加载成功"
else
    echo "   ❌ 前端首页加载失败"
fi

echo ""
echo "2. 测试登录功能"
echo "   - 尝试登录 admin/admin123"
LOGIN_RESPONSE=$(curl -s -X POST $API_URL/auth/login -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}')
if echo "$LOGIN_RESPONSE" | grep -q "token"; then
    echo "   ✅ 登录成功"
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    echo "   - Token: ${TOKEN:0:20}..."
else
    echo "   ❌ 登录失败"
fi

echo ""
echo "3. 测试API功能"

echo "   - 获取链接列表"
LINKS_RESPONSE=$(curl -s $API_URL/links)
if echo "$LINKS_RESPONSE" | grep -q "abc123"; then
    echo "   ✅ 链接列表获取成功"
    LINK_COUNT=$(echo "$LINKS_RESPONSE" | grep -o "link_id" | wc -l)
    echo "   - 找到 $LINK_COUNT 个链接"
else
    echo "   ❌ 链接列表获取失败"
fi

echo ""
echo "   - 获取系统统计"
STATS_RESPONSE=$(curl -s $API_URL/stats/system)
if echo "$STATS_RESPONSE" | grep -q "total_links"; then
    echo "   ✅ 系统统计获取成功"
    echo "   - 响应: $STATS_RESPONSE"
else
    echo "   ❌ 系统统计获取失败"
fi

echo ""
echo "4. 测试重定向功能"
echo "   - 测试URL: /api/v1/redirect/bu01/abc123?network=mi&kw=test"
REDIRECT_RESPONSE=$(curl -s -I "$FRONTEND_URL/api/v1/redirect/bu01/abc123?network=mi&kw=test")
if echo "$REDIRECT_RESPONSE" | grep -q "302 Found"; then
    echo "   ✅ 重定向功能正常"
    LOCATION=$(echo "$REDIRECT_RESPONSE" | grep "Location:" | cut -d' ' -f2)
    echo "   - 重定向到: $LOCATION"
else
    echo "   ❌ 重定向功能失败"
fi

echo ""
echo "=== 测试完成 ==="