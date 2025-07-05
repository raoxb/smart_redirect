#!/usr/bin/env python3
"""
为短链添加多条目标链接进行测试
"""

import psycopg2
import json
from datetime import datetime

def add_test_targets():
    """为现有链接添加测试目标"""
    
    # 连接数据库
    conn = psycopg2.connect(
        host='localhost',
        database='smart_redirect',
        user='smartredirect',
        password='smart123'
    )
    cur = conn.cursor()
    
    print("🎯 为短链添加测试目标...")
    
    # 为 abc123 (link_id=1) 添加更多目标
    abc123_targets = [
        {
            'url': 'https://example.com/landing1',
            'weight': 30,
            'cap': 1000,
            'countries': '["US", "CA"]',
            'param_mapping': '{"utm_source": "src", "utm_campaign": "camp"}',
            'static_params': '{"ref": "test1"}'
        },
        {
            'url': 'https://example.com/landing2', 
            'weight': 25,
            'cap': 800,
            'countries': '["UK", "DE", "FR"]',
            'param_mapping': '{"keyword": "kw"}',
            'static_params': '{"ref": "test2"}'
        },
        {
            'url': 'https://example.com/landing3',
            'weight': 20,
            'cap': 600,
            'countries': '["AU", "NZ"]',
            'param_mapping': '{}',
            'static_params': '{"ref": "test3", "version": "v2"}'
        },
        {
            'url': 'https://example.com/fallback',
            'weight': 15,
            'cap': 0,  # 无限制
            'countries': '["ALL"]',
            'param_mapping': '{"q": "query"}',
            'static_params': '{"ref": "fallback"}'
        }
    ]
    
    # 为 def456 (link_id=2) 添加更多目标
    def456_targets = [
        {
            'url': 'https://example.com/promo1',
            'weight': 40,
            'cap': 1200,
            'countries': '["US"]',
            'param_mapping': '{"source": "s", "medium": "m"}',
            'static_params': '{"promo": "summer2024"}'
        },
        {
            'url': 'https://example.com/promo2',
            'weight': 35,
            'cap': 900,
            'countries': '["CN", "JP", "KR"]',
            'param_mapping': '{}',
            'static_params': '{"promo": "asia", "lang": "zh"}'
        },
        {
            'url': 'https://example.com/global',
            'weight': 25,
            'cap': 0,  # 无限制
            'countries': '["ALL"]',
            'param_mapping': '{"utm_content": "content"}',
            'static_params': '{"global": "true"}'
        }
    ]
    
    # 插入 abc123 的目标
    link_id_1 = 1
    for target in abc123_targets:
        cur.execute("""
            INSERT INTO targets (link_id, url, weight, cap, current_hits, countries, param_mapping, static_params, is_active, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
        """, (
            link_id_1,
            target['url'], 
            target['weight'],
            target['cap'],
            0,  # current_hits
            target['countries'],
            target['param_mapping'],
            target['static_params'],
            True,  # is_active
            datetime.now(),
            datetime.now()
        ))
        print(f"  ✅ 为 abc123 添加目标: {target['url']} (权重: {target['weight']}%)")
    
    # 插入 def456 的目标
    link_id_2 = 2
    for target in def456_targets:
        cur.execute("""
            INSERT INTO targets (link_id, url, weight, cap, current_hits, countries, param_mapping, static_params, is_active, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
        """, (
            link_id_2,
            target['url'],
            target['weight'], 
            target['cap'],
            0,  # current_hits
            target['countries'],
            target['param_mapping'],
            target['static_params'],
            True,  # is_active
            datetime.now(),
            datetime.now()
        ))
        print(f"  ✅ 为 def456 添加目标: {target['url']} (权重: {target['weight']}%)")
    
    # 提交更改
    conn.commit()
    
    # 验证结果
    cur.execute("SELECT link_id, COUNT(*) FROM targets GROUP BY link_id ORDER BY link_id")
    results = cur.fetchall()
    
    print(f"\n📊 目标统计:")
    for link_id, count in results:
        cur.execute("SELECT link_id FROM links WHERE id = %s", (link_id,))
        link_code = cur.fetchone()[0]
        print(f"  链接 {link_code}: {count} 个目标")
    
    # 显示权重分布
    print(f"\n⚖️  权重分布:")
    cur.execute("""
        SELECT l.link_id, t.url, t.weight, t.countries 
        FROM targets t 
        JOIN links l ON t.link_id = l.id 
        ORDER BY l.link_id, t.weight DESC
    """)
    targets = cur.fetchall()
    
    current_link = None
    for link_code, url, weight, countries in targets:
        if current_link != link_code:
            current_link = link_code
            print(f"\n  📎 {link_code}:")
        
        # 解析国家列表
        try:
            country_list = json.loads(countries) if countries else []
            country_str = ', '.join(country_list) if country_list else 'All'
        except:
            country_str = 'All'
            
        print(f"    - {url} ({weight}%) [{country_str}]")
    
    cur.close()
    conn.close()
    
    print(f"\n🎉 测试目标添加完成!")
    print(f"现在可以测试多目标权重分配和地域定向功能了。")

if __name__ == "__main__":
    add_test_targets()