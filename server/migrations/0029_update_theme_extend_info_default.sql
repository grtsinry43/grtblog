-- +goose Up
INSERT INTO website_info (info_key, name, value, info_json)
VALUES (
    'theme_extend_info',
    '主题扩展信息',
    NULL,
    $json$
{
  "home": {
    "hero": {
      "title": {
        "template": [
          {
            "text": "Hi! 👋",
            "type": "h1",
            "variant": "hero_h1_highlight"
          },
          {
            "type": "br"
          },
          {
            "text": "I'm grtsinry43",
            "type": "h1",
            "variant": "hero_h1_primary"
          }
        ]
      },
      "socials": [
        {
          "href": "https://github.com/grtinry43",
          "icon": "github",
          "name": "GitHub"
        },
        {
          "href": "mailto:grtsinry43@outlook.com",
          "icon": "mail",
          "name": "Email"
        },
        {
          "href": "/feed",
          "icon": "rss",
          "name": "RSS"
        }
      ],
      "avatarUrl": "https://dogeoss.grtsinry43.com/img/author.jpeg",
      "mottoLines": [
        "热衷于在逻辑与感性的缝隙中构建数字花园。",
        "也许，代码是现代的诗歌，而文字是思想的快照。"
      ],
      "description": "Java & JavaScript full-stack developer committed to crafting excellent software."
    },
    "inspiration": {
      "now": {
        "items": [
          {
            "id": "coding",
            "icon": "code2",
            "label": "Coding",
            "value": "grtblog-v2"
          },
          {
            "id": "reading",
            "icon": "library",
            "label": "Reading",
            "value": "The Design of Everyday Things"
          },
          {
            "id": "learning",
            "icon": "zap",
            "label": "Learning",
            "value": "Svelte 5 & Runes"
          }
        ],
        "title": "Now / 正在"
      },
      "quote": {
        "text": "“Nothing but enthusiasm brightens up the endless years.”",
        "author": ""
      },
      "stats": [
        {
          "id": "words",
          "icon": "library",
          "label": "Words",
          "source": {
            "type": "words_total"
          },
          "colorClass": "text-jade-500"
        },
        {
          "id": "commits",
          "icon": "github",
          "label": "Commits",
          "source": {
            "type": "github_recent_push_commits"
          },
          "colorClass": "text-ink-900 dark:text-ink-100"
        },
        {
          "id": "coffee",
          "icon": "coffee",
          "label": "Coffee",
          "value": "∞",
          "source": {
            "type": "static"
          },
          "colorClass": "text-amber-500"
        }
      ],
      "energy": {
        "icon": "sparkles",
        "label": "High Energy"
      },
      "github": {
        "username": "grtsinry43"
      },
      "techStack": {
        "icons": [
          "code2",
          "gamepad2"
        ],
        "items": [
          "Java",
          "TypeScript",
          "Svelte",
          "Rust"
        ],
        "title": "Tech Stack"
      },
      "sectionTitle": "灵感与实验场"
    },
    "activityPulse": {
      "title": "创作律动",
      "legend": {
        "posts": "Article",
        "moments": "Moment"
      },
      "subtitle": "近一年的数字足迹：逻辑的向上生长，感性的向下扎根。",
      "rangeDays": 365,
      "rangeLabelEnd": "Today",
      "rangeLabelStart": "365 Days Ago"
    }
  }
}
$json$::jsonb
)
ON CONFLICT (info_key) DO UPDATE
SET name = EXCLUDED.name,
    info_json = EXCLUDED.info_json,
    updated_at = now();

-- +goose Down
INSERT INTO website_info (info_key, name, value, info_json)
VALUES ('theme_extend_info', '主题扩展信息', NULL, '{}'::jsonb)
ON CONFLICT (info_key) DO UPDATE
SET name = EXCLUDED.name,
    info_json = EXCLUDED.info_json,
    updated_at = now();
