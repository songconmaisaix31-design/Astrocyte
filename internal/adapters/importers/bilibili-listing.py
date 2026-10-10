"""One credential-free metadata page via the installed pinned upstream extractor.

The query follows BilibiliSpaceVideoIE; signing/fingerprints/networking remain
upstream yt-dlp. No config files, browser cookies, video extraction or downloads.
"""
import json
import re
import sys

from yt_dlp import YoutubeDL
from yt_dlp.extractor.bilibili import BilibiliSpaceVideoIE
from yt_dlp.version import __version__

uid, page, size = sys.argv[1:]
if __version__ != "2026.08.19" or not re.fullmatch(r"[1-9][0-9]{0,19}", uid):
    raise ValueError("pinned yt-dlp and positive public UID required")
page = int(page)
size = int(size)
if not 1 <= page <= 3000000 or not 1 <= size <= 30:
    raise ValueError("bounded page required")

with YoutubeDL({"quiet": True, "no_warnings": True, "socket_timeout": 20,
                "retries": 0, "extractor_retries": 0, "skip_download": True,
                "cachedir": False, "cookiefile": None,
                "cookiesfrombrowser": None}) as ydl:
    ie = BilibiliSpaceVideoIE(ydl)
    query = {"keyword": "", "mid": uid, "order": "pubdate",
             "order_avoided": "true", "platform": "web", "pn": page,
             "ps": size, "tid": 0, "web_location": "333.1387",
             "special_type": "", "index": 0, **ie._dm_params}
    result = ie._download_json(
        "https://api.bilibili.com/x/space/wbi/arc/search", uid,
        query=ie._sign_wbi(query, uid),
        headers={"Referer": f"https://space.bilibili.com/{uid}/video",
                 "Origin": "https://space.bilibili.com",
                 "Accept-Language": "en,zh-CN;q=0.9,zh;q=0.8"})
    print(json.dumps(result, ensure_ascii=False))
