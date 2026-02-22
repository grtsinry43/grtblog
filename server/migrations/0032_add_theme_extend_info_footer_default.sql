-- +goose Up
INSERT INTO website_info (info_key, name, value, info_json)
VALUES (
    'theme_extend_info',
    '主题扩展信息',
    NULL,
    $json$
{
  "footer": {
    "sections": [
      {
        "title": "想要了解我",
        "links": [
          { "name": "关于我", "href": "/about" },
          { "name": "本站历史", "href": "/about-site" },
          { "name": "关于此项目", "href": "/about-project" }
        ]
      },
      {
        "title": "你也许在找",
        "links": [
          { "name": "归档", "href": "/posts" },
          { "name": "友链", "href": "/friends" },
          { "name": "RSS", "href": "/feed" },
          { "name": "时间线", "href": "/timeline" },
          { "name": "监控", "href": "https://status.grtsinry43.com" }
        ]
      },
      {
        "title": "联系我叭",
        "links": [
          { "name": "写留言", "href": "/message" },
          { "name": "发邮件", "href": "mailto:grtsinry43@outlook.com" },
          { "name": "GitHub", "href": "https://github.com/grtsinry43" }
        ]
      }
    ],
    "brand": {
      "name": "Grtsinry43's Blog.",
      "tagline": "总之岁月漫长，然而值得等待"
    },
    "copyright": {
      "startYear": 2022,
      "owner": "grtsinry43",
      "beianText": "",
      "beianUrl": "https://beian.miit.gov.cn/",
      "beianGongAnText": "",
      "designedWithText": "Designed by Grtsinry43 with ❤"
    },
    "presence": {
      "connectedText": "正在有 {count} 位小伙伴看着我的网站呐",
      "loadingText": "正在同步在线状态..."
    }
  }
}
$json$::jsonb
)
ON CONFLICT (info_key) DO UPDATE
SET name = EXCLUDED.name,
    info_json = CASE
        WHEN jsonb_typeof(website_info.info_json) = 'object'
             AND jsonb_typeof(website_info.info_json -> 'footer') = 'object' THEN website_info.info_json
        WHEN jsonb_typeof(website_info.info_json) = 'object' THEN website_info.info_json || EXCLUDED.info_json
        ELSE EXCLUDED.info_json
    END,
    updated_at = now();

-- +goose Down
UPDATE website_info
SET info_json = CASE
    WHEN jsonb_typeof(info_json) = 'object' THEN info_json - 'footer'
    ELSE '{}'::jsonb
END,
updated_at = now()
WHERE info_key = 'theme_extend_info';
