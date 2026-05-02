#!/usr/bin/env python3
"""
获取股票诉讼仲裁及违规担保信息
数据来源：巨潮资讯网(cninfo.com.cn) 公告查询
输入：{"symbol": "000001.SZ"}
输出：{"litigation_count": 0, "has_litigation": false, "litigation_amount": 0, "history": [...]}
"""
import json
import sys
import os
import re
from datetime import datetime, timedelta

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, SCRIPT_DIR)
from cninfo_utils import query_announcements, filter_by_keywords


# 诉讼/仲裁/担保/违规相关关键词
LITIGATION_KEYWORDS = [
    "诉讼", "仲裁", "起诉", "被诉", "涉诉", "纠纷案件",
    "合同纠纷", "债务纠纷", "股权纠纷", "侵权纠纷",
]
# 高风险担保关键词（一票否决级别）
HIGH_RISK_GUARANTEE_KEYWORDS = [
    "违规担保", "担保逾期", "担保代偿", "承担担保责任",
    "担保追偿", "无真实交易背景担保", "为控股股东提供担保",
    "为实际控制人提供担保",
]
# 一般担保关键词（中风险累积）
GUARANTEE_KEYWORDS = [
    "对外担保", "关联担保", "担保进展", "反担保",
]
VIOLATION_KEYWORDS = [
    "违规", "处罚", "立案调查", "警示函", "监管函",
    "责令改正", "行政处罚", "市场禁入",
]
# 排除关键词（正常经营活动，不计入风险）
EXCLUDE_KEYWORDS = [
    "担保额度", "预计担保", "年度担保",
    "为控股子公司提供担保", "为全资子公司提供担保",
    "为子公司提供担保", "股东大会审议担保", "董事会审议担保",
]
ALL_KEYWORDS = LITIGATION_KEYWORDS + HIGH_RISK_GUARANTEE_KEYWORDS + GUARANTEE_KEYWORDS + VIOLATION_KEYWORDS


def extract_amount(title: str) -> float:
    """尝试从标题中提取涉诉金额（万元）"""
    # 匹配模式如：涉及金额 1234.56 万元 / 涉案金额约 1.2 亿元
    patterns = [
        r'涉及金额[约]*\s*(\d+\.?\d*)\s*万元',
        r'涉案金额[约]*\s*(\d+\.?\d*)\s*万元',
        r'标的金额[约]*\s*(\d+\.?\d*)\s*万元',
        r'(\d+\.?\d*)\s*万元',
    ]
    for p in patterns:
        m = re.search(p, title)
        if m:
            return float(m.group(1))
    return 0.0


def fetch_litigation(symbol: str):
    """获取股票诉讼仲裁及违规担保信息"""
    try:
        start = (datetime.now() - timedelta(days=365 * 2)).strftime("%Y-%m-%d")

        # 查询全部公告，然后筛选关键词
        announcements = query_announcements(
            symbol, search_key="", start_date=start, page_size=200
        )

        if announcements and "error" in announcements[0]:
            return {
                "litigation_count": 0,
                "has_litigation": False,
                "litigation_amount": 0,
                "history": [],
                "error": announcements[0]["error"],
            }

        filtered = filter_by_keywords(announcements, ALL_KEYWORDS)

        # 排除正常经营活动公告
        risk_items = []
        for a in filtered:
            title = a.get("announcementTitle", "")
            # 排除正常担保额度类公告
            if any(kw in title for kw in EXCLUDE_KEYWORDS):
                continue
            risk_items.append(a)

        total_amount = 0.0
        has_guarantee = False
        has_high_risk_guarantee = False
        has_fund_occupation = False
        high_risk_history = []
        medium_risk_history = []
        for a in risk_items:
            title = a.get("announcementTitle", "")
            date = a.get("announcementDate", "")
            amount = extract_amount(title)
            total_amount += amount

            # 分类并判断风险等级
            category = "其他"
            is_high = False
            for kw in LITIGATION_KEYWORDS:
                if kw in title:
                    category = "诉讼仲裁"
                    is_high = True
                    break
            if category == "其他":
                for kw in HIGH_RISK_GUARANTEE_KEYWORDS:
                    if kw in title:
                        category = "高风险担保"
                        has_high_risk_guarantee = True
                        has_guarantee = True
                        is_high = True
                        break
            if category == "其他":
                for kw in GUARANTEE_KEYWORDS:
                    if kw in title:
                        category = "担保"
                        has_guarantee = True
                        break
            if category == "其他":
                for kw in VIOLATION_KEYWORDS:
                    if kw in title:
                        category = "违规处罚"
                        is_high = True
                        break
            # 资金占用检测（一票否决级别）
            if "资金占用" in title or "占用资金" in title:
                has_fund_occupation = True
                is_high = True

            item = f"[{date}] [{category}] {title}"
            if is_high:
                high_risk_history.append(item)
            else:
                medium_risk_history.append(item)

        # 高风险在前，中风险在后
        history = high_risk_history + medium_risk_history

        return {
            "litigation_count": len(high_risk_history),
            "has_guarantee": has_guarantee,
            "has_high_risk_guarantee": has_high_risk_guarantee,
            "has_fund_occupation": has_fund_occupation,
            "history": history[:10],  # 最多返回10条
        }

    except Exception as e:
        return {
            "litigation_count": 0,
            "has_litigation": False,
            "litigation_amount": 0,
            "history": [],
            "error": str(e),
        }


def main():
    req = json.load(sys.stdin)
    symbol = req.get("symbol", "")
    result = fetch_litigation(symbol)
    print(json.dumps(result, ensure_ascii=False))


if __name__ == "__main__":
    main()
