#!/usr/bin/env python3
"""
获取股票高管变动信息
数据来源：巨潮资讯网(cninfo.com.cn) 公告查询
输入：{"symbol": "000001.SZ"}
输出：{"exec_change_count": 0, "cfo_changed": false, "audit_head_changed": false, "history": [...]}
"""
import json
import sys
import os
import re
from datetime import datetime, timedelta

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, SCRIPT_DIR)
from cninfo_utils import query_announcements, filter_by_keywords


# 识别 CFO / 财务总监 / 审计负责人变动的关键词
CFO_KEYWORDS = [
    "财务总监", "总会计师", "首席财务官", "CFO", "财务负责人",
    "分管财务", "财务工作",
]
AUDIT_HEAD_KEYWORDS = [
    "审计部负责人", "审计负责人", "内部审计负责人",
    "审计委员会委员", "审计委员会主任",
]


def is_cfo_related(title: str) -> bool:
    """判断公告是否涉及 CFO/财务总监变动"""
    t = title.lower()
    for kw in CFO_KEYWORDS:
        if kw in t:
            return True
    return False


def is_audit_head_related(title: str) -> bool:
    """判断公告是否涉及审计负责人变动"""
    for kw in AUDIT_HEAD_KEYWORDS:
        if kw in title:
            return True
    return False


def fetch_exec_changes(symbol: str):
    """获取股票高管变动信息"""
    try:
        start = (datetime.now() - timedelta(days=365)).strftime("%Y-%m-%d")

        # 分别查询 CFO/审计/高管相关公告，合并去重
        # 优先用 CFO/审计专属关键词搜索，减少初始噪音
        ann_cfo = query_announcements(
            symbol, search_key="财务总监", start_date=start, page_size=30
        )
        ann_cfo2 = query_announcements(
            symbol, search_key="CFO", start_date=start, page_size=30
        )
        ann_shenji = query_announcements(
            symbol, search_key="审计部", start_date=start, page_size=30
        )
        ann_gaoguan = query_announcements(
            symbol, search_key="高级管理人员", start_date=start, page_size=30
        )
        ann_dongshi = query_announcements(
            symbol, search_key="董事", start_date=start, page_size=30
        )

        # 合并去重（按公告标题+日期）
        seen = set()
        all_ann = []
        for a in ann_cfo + ann_cfo2 + ann_shenji + ann_gaoguan + ann_dongshi:
            key = a.get("announcementTitle", "") + a.get("announcementDate", "")
            if key and key not in seen:
                seen.add(key)
                all_ann.append(a)

        if all_ann and "error" in all_ann[0]:
            return {
                "exec_change_count": 0,
                "cfo_changed": False,
                "audit_head_changed": False,
                "history": [],
                "error": all_ann[0]["error"],
            }

        # 只保留标题明确涉及人事变动的公告
        # 要求标题同时满足：(1)含"董事/高级管理人员"相关词  (2)含变动相关词
        CHANGE_KEYWORDS = ["变动", "辞职", "离任", "聘任", "任免", "解聘", "变更",
                           "调整", "选举", "增补", "补选", "换届", "辞任", "免职"]
        EXCLUDE_KEYWORDS = ["薪酬", "考核", "述职", "制度", "章程", "议事规则",
                            "工作细则", "会议决议", "独立性", "任职资格核准",
                            "履职情况", "评估报告"]
        filtered = []
        for a in all_ann:
            title = a.get("announcementTitle", "")
            # 排除明显无关
            if any(kw in title for kw in EXCLUDE_KEYWORDS):
                continue
            # 必须包含变动关键词
            if not any(kw in title for kw in CHANGE_KEYWORDS):
                continue
            filtered.append(a)

        cfo_changed = False
        audit_head_changed = False
        history = []
        for a in filtered:
            title = a.get("announcementTitle", "")
            date = a.get("announcementDate", "")
            is_cfo = is_cfo_related(title)
            is_audit = is_audit_head_related(title)
            if not (is_cfo or is_audit):
                continue  # 只保留财务/审计负责人相关变动
            if is_cfo:
                cfo_changed = True
            if is_audit:
                audit_head_changed = True
            tags = []
            if is_cfo:
                tags.append("CFO")
            if is_audit:
                tags.append("审计")
            tag_str = f" [{'/'.join(tags)}]" if tags else ""
            history.append(f"[{date}]{tag_str} {title}")

        return {
            "exec_change_count": len(history),
            "cfo_changed": cfo_changed,
            "audit_head_changed": audit_head_changed,
            "history": history[:10],  # 最多返回10条
        }

    except Exception as e:
        return {
            "exec_change_count": 0,
            "cfo_changed": False,
            "audit_head_changed": False,
            "history": [],
            "error": str(e),
        }


def main():
    req = json.load(sys.stdin)
    symbol = req.get("symbol", "")
    result = fetch_exec_changes(symbol)
    print(json.dumps(result, ensure_ascii=False))


if __name__ == "__main__":
    main()
