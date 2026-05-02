#!/usr/bin/env python3
"""
巨潮资讯网(cninfo.com.cn) 公告查询公共工具
供 fetch_auditor_history.py / fetch_exec_changes.py / fetch_litigation.py 共用
"""
import requests
import json
from datetime import datetime, timedelta


def get_org_id(symbol: str) -> tuple:
    """
    通过股票代码查询巨潮资讯网的 orgId 和交易所(column)
    返回: (orgId, column, plate)
    """
    pure_code = symbol.split('.')[0]
    url = "http://www.cninfo.com.cn/new/information/topSearch/query"
    try:
        resp = requests.post(url, data={"keyWord": pure_code}, timeout=10)
        data = resp.json()
        if isinstance(data, list) and len(data) > 0:
            item = data[0]
            org_id = item.get("orgId", "")
            # column: szse(深交所), sse(上交所), bse(北交所)
            column = "szse"
            plate = "sz"
            if org_id.startswith("gssh"):
                column = "sse"
                plate = "sh"
            elif org_id.startswith("gsbj"):
                column = "bse"
                plate = "bj"
            return org_id, column, plate
    except Exception:
        pass
    # fallback: 按代码规则推断
    if pure_code.startswith(("6", "68", "69")):
        return f"gssh{pure_code}", "sse", "sh"
    elif pure_code.startswith(("8", "4", "43")):
        return f"gsbj{pure_code}", "bse", "bj"
    else:
        return f"gssz{pure_code}", "szse", "sz"


def query_announcements(
    symbol: str,
    search_key: str = "",
    start_date: str = "",
    end_date: str = "",
    page_size: int = 100,
) -> list:
    """
    查询巨潮资讯网公告列表
    返回: [{announcementTitle, announcementTime, adjunctUrl, ...}, ...]
    """
    pure_code = symbol.split('.')[0]
    org_id, column, _ = get_org_id(symbol)

    if not end_date:
        end_date = datetime.now().strftime("%Y-%m-%d")
    if not start_date:
        start_date = (datetime.now() - timedelta(days=365)).strftime("%Y-%m-%d")

    url = "http://www.cninfo.com.cn/new/hisAnnouncement/query"
    payload = {
        "pageNum": "1",
        "pageSize": str(page_size),
        "column": column,
        "tabName": "fulltext",
        "plate": "",
        "stock": f"{pure_code},{org_id}",
        "searchkey": search_key,
        "secid": "",
        "category": "",
        "trade": "",
        "seDate": f"{start_date}~{end_date}",
        "sortName": "",
        "sortType": "",
    }
    try:
        resp = requests.post(url, params=payload, timeout=15)
        data = resp.json()
        announcements = data.get("announcements") or []
        # 转换时间戳为日期字符串
        for a in announcements:
            ts = a.get("announcementTime", 0)
            if ts:
                a["announcementDate"] = datetime.fromtimestamp(
                    ts / 1000
                ).strftime("%Y-%m-%d")
            else:
                a["announcementDate"] = ""
        return announcements
    except Exception as e:
        return [{"error": str(e)}]


def filter_by_keywords(announcements: list, keywords: list) -> list:
    """筛选标题包含任意关键词的公告"""
    result = []
    for a in announcements:
        title = a.get("announcementTitle", "")
        for kw in keywords:
            if kw in title:
                result.append(a)
                break
    return result


def extract_year_from_title(title: str) -> str:
    """从公告标题中提取年份，如'2024年年度报告' -> '2024'"""
    import re
    m = re.search(r'(20\d{2})', title)
    return m.group(1) if m else ""
