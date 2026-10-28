# 200t.vinted.mi 接口字段

## 协议身份

- 插件 ID：`200t.vinted.mi`。
- Provider ID：`200t-vinted-mi`。
- 能力：`video`。
- 默认 Base URL：`https://vinted.cam`。
- 鉴权驱动：`bearer`。
- 创建：`POST /v1/videos`。
- 查询：`GET /v1/videos/{{taskId}}`。

## 配置字段

| 字段 | 类型 | 必填 | 含义 |
| --- | --- | --- | --- |
| `apiKey` | secret | 是 | API Key |

## 统一字段映射

| 统一字段 | 类型 | 必填 | 上游映射 | 说明 |
| --- | --- | --- | --- | --- |
| `model` | string | 是 | `model_id` | 仅支持 multiref:sd-2-c4、multiref:ali-wan3、qsqs:wan3-720p-deal。 |
| `prompt` | string | 是 | `prompt/content/input` | 视频提示词。 |
| `images` | media[] | 否 | `first/last/reference image` | 显式 role 图片输入。 |
| `videos` | media[] | 否 | `reference video` | 参考视频。 |
| `audios` | media[] | 否 | `reference audio/voice` | 参考音频或音色。 |
| `duration` | integer | 否 | `duration/seconds` | 时长秒数。 |
| `aspectRatio` | string | 否 | `ratio/aspect_ratio/size` | 画幅比例或尺寸。 |
| `resolution` | string | 否 | `resolution` | 分辨率档位。 |
| `generateAudio` | boolean | 否 | `generate_audio` | 是否生成音频。 |
| `providerOptions` | object | 否 | `provider-specific fields` | 插件命名空间内的厂商扩展字段。 |

## 上游请求模板逐字段清单

下表由插件请求模板生成，覆盖 body、query、headers 和 multipart 文件声明中的每个字段。

| 上游位置 | 值或转换表达式 |
| --- | --- |
| `create.method` | `"POST"` |
| `create.path` | `"/v1/videos"` |
| `create.contentType` | `"application/json"` |
| `create.body` | `{"$ref":"request.extra.preparation.quote.body"}` |
| `create.headers.Idempotency-Key` | `{"$ref":"request.extra.idempotencyKey"}` |
| `create.headers.Expected-Quote-Id` | `{"$ref":"request.extra.preparation.quote.id"}` |
| `poll.method` | `"GET"` |
| `poll.path` | `"/v1/videos/{{taskId}}"` |
| `poll.contentType` | `"application/json"` |
| `result.method` | `"GET"` |
| `result.path` | `"/v1/videos/{{taskId}}/signed_url"` |
| `result.contentType` | `"application/json"` |

## Provider 扩展键

- 无额外扩展键。

动态模型或工作流允许使用文档声明的完整 `parameters/input/extra_body` 对象；该对象是协议本身的开放 schema，不会被宿主裁剪。

## 响应映射逐字段清单

| 映射位置 | 上游路径或转换表达式 |
| --- | --- |
| `response.taskId` | `{"$coalesce":[{"$ref":"taskId"},{"$ref":"response.id"}]}` |
| `response.status` | `{"$if":{"condition":{"$and":[{"$ref":"taskId"},{"$ref":"response.id"},{"$ne":[{"$ref":"response.id"},{"$ref":"taskId"}]}]},"then":"failed","else":{"$if":{"condition":{"$and":[{"$not":{"$ref":"response.status"}},{"$coalesce":[{"$ref":"response.url"},{"$ref":"response.download_url"}]}]},"then":"succeeded","else":{"$if":{"condition":{"$and":[{"$or":[{"$eq":[{"$ref":"response.status"},"succeeded"]},{"$and":[{"$eq":[{"$ref":"response.status"},"failed"]},{"$eq":[{"$ref":"response.refunded"},true]},{"$eq":[{"$ref":"response.has_file"},true]},{"$eq":[{"$ref":"response.recovered_artifact.available"},true]},{"$eq":[{"$ref":"response.recovered_artifact.settlement_unchanged"},true]}]}]},{"$eq":[{"$ref":"response.has_file"},true]}]},"then":"succeeded","else":{"$if":{"condition":{"$eq":[{"$ref":"response.status"},"failed"]},"then":"failed","else":{"$if":{"condition":{"$eq":[{"$ref":"response.status"},"queued"]},"then":"pending","else":"processing"}}}}}}}}}}` |
| `response.videos` | `{"$if":{"condition":{"$not":{"$ref":"response.status"}},"then":{"$coalesce":[{"$ref":"response.url"},{"$ref":"response.download_url"}]},"else":null}}` |
| `response.message` | `{"$coalesce":[{"$ref":"response.error.message"},{"$ref":"response.error"},{"$ref":"response.progress.message"}]}` |
| `response.usage` | `{"$ref":"response.result.actual"}` |
| `response.resultEphemeral` | `true` |

## 响应与错误

插件把上游 task/status/text/media/usage 映射为统一结果。临时媒体 URL 标记为 ephemeral，由宿主立即下载持久化。HTTP 错误、业务 code 和 error object 保持失败语义，不包装成成功。

## 兼容边界

接入文档：https://vinted.cam/developers 。仅支持 C4、阿里满血 Wan3、Wan3 720p Deal。

三个模型均提供业务层文生视频。C4 和阿里 Wan3 的当前 input_schema 仅要求 prompt，未要求参考素材；无素材时仍使用上游 input_mode=references，而不是 text。Deal 无素材使用 text，有素材使用 references。不能根据 references 模式名推断必须上传素材；模式、参数与组合最终由实时目录和报价接口确认。

C4 支持 4–15 整数秒、720p、9 图/3 视频/3 音频；图片宽高须大于 300，参考音频合计不超过 15 秒，视频合计加成片不超过 25 秒。首尾帧使用 frames，images 按首帧、尾帧排序，不混音视频。阿里 Wan3 支持 4–30 整数秒、480p/720p/1080p、10 图/5 视频/5 音频；视频与音频各合计不超过 15 秒，视频合计加成片不超过 30 秒；声音开关映射 save_audio。Deal 仅 5/10/15/30 秒、720p、16:9/9:16、10 图/5 音频，禁止视频参考，不发送 save_audio。UI 默认能力为已知规格。

每次生成读取目录，要求 ready、有效 revision、未过期 valid_until 和 available=true。素材按 x-canvas-inputs 映射，同字段合并有序引用；参数仅发送 parameter_schema.properties 声明的键，支持 quality/resolution 与显式默认值。高级扩展放入 providerOptions.200t-vinted-mi.parameters，不能覆盖已选择并用于计费的统一参数。素材先通过带名称的 /v1/files 获取 asset_id，不发送公网地址或旧 image_id；图库容量由客户管理。

完整请求先报价，再用相同正文、Expected-Quote-Id 和固定 Idempotency-Key 提交；价格变化不自动换报价。创建重试最多三次，正文及编号不变。无任务 id 但结果不明时，使用原 Key 与日志中的操作编号查询 /v1/submissions/{key}，不新建订单；有 id 后仅查询原任务。按次计费，不乘秒数。轮询检查 has_file 与补交标记，signed_url 使用独立匿名请求下载 MP4，不依赖旧质量响应头。

<!-- YINGCE_MANIFEST_CONTRACT_START -->
## Manifest 完整接口定义

以下 JSON 与插件包内实际 `manifest.json` 逐字段一致，覆盖插件身份、权限、配置、鉴权、参数、校验、创建、Agent、查询、取消、结果下载、响应和 Agent 响应映射。`documentation` 字段的值就是当前完整文档；为避免文档在自身内部无限递归，JSON 中仅用等义占位文本表示正文。

```json
{
  "apiVersion": "yingce.plugin/v2",
  "runtime": {
    "backend": "declarative-lifecycle"
  },
  "id": "200t.vinted.mi",
  "name": "200t.vinted.mi",
  "version": "1.2.0",
  "author": "200t / Vinted / 影策",
  "description": "200t.vinted.mi 独立请求协议插件。",
  "documentation": "<当前插件的完整 documentation，由 README.md 与 docs/interface.md 拼接而成；为避免 JSON 递归，此处不重复展开正文。>",
  "permissions": [
    "generation.run",
    "media.read"
  ],
  "configuration": {
    "fields": [
      {
        "name": "apiKey",
        "type": "secret",
        "label": "API Key",
        "required": true
      }
    ]
  },
  "contributes": {
    "providers": [
      {
        "id": "200t-vinted-mi",
        "label": "200t.vinted.mi",
        "capabilities": [
          "video"
        ],
        "scopes": [
          "admin.system-channel",
          "user.custom-channel",
          "canvas",
          "creation",
          "agent"
        ],
        "baseUrl": "https://vinted.cam",
        "requiresPublicMediaUrls": false,
        "auth": {
          "type": "bearer",
          "field": "apiKey"
        },
        "parameters": [
          {
            "name": "model",
            "type": "string",
            "required": true,
            "mapping": "model_id",
            "description": "仅支持 multiref:sd-2-c4、multiref:ali-wan3、qsqs:wan3-720p-deal。"
          },
          {
            "name": "prompt",
            "type": "string",
            "required": true,
            "mapping": "prompt/content/input",
            "description": "视频提示词。"
          },
          {
            "name": "images",
            "type": "media[]",
            "required": false,
            "mapping": "first/last/reference image",
            "description": "显式 role 图片输入。"
          },
          {
            "name": "videos",
            "type": "media[]",
            "required": false,
            "mapping": "reference video",
            "description": "参考视频。"
          },
          {
            "name": "audios",
            "type": "media[]",
            "required": false,
            "mapping": "reference audio/voice",
            "description": "参考音频或音色。"
          },
          {
            "name": "duration",
            "type": "integer",
            "required": false,
            "mapping": "duration/seconds",
            "description": "时长秒数。"
          },
          {
            "name": "aspectRatio",
            "type": "string",
            "required": false,
            "mapping": "ratio/aspect_ratio/size",
            "description": "画幅比例或尺寸。"
          },
          {
            "name": "resolution",
            "type": "string",
            "required": false,
            "mapping": "resolution",
            "description": "分辨率档位。"
          },
          {
            "name": "generateAudio",
            "type": "boolean",
            "required": false,
            "mapping": "generate_audio",
            "description": "是否生成音频。"
          },
          {
            "name": "providerOptions",
            "type": "object",
            "required": false,
            "mapping": "provider-specific fields",
            "description": "插件命名空间内的厂商扩展字段。"
          }
        ],
        "validations": [
          {
            "assert": {
              "$in": [
                {
                  "$ref": "request.model"
                },
                [
                  "multiref:sd-2-c4",
                  "multiref:ali-wan3",
                  "qsqs:wan3-720p-deal"
                ]
              ]
            },
            "message": "200t.vinted.mi 仅支持 C4、阿里 Wan3 和 Wan3 720p Deal"
          },
          {
            "assert": {
              "$gt": [
                {
                  "$len": {
                    "$trim": {
                      "$ref": "request.prompt"
                    }
                  }
                },
                0
              ]
            },
            "message": "视频提示词不能为空"
          },
          {
            "assert": {
              "$if": {
                "condition": {
                  "$eq": [
                    {
                      "$ref": "request.model"
                    },
                    "qsqs:wan3-720p-deal"
                  ]
                },
                "then": {
                  "$eq": [
                    {
                      "$len": {
                        "$ref": "request.videos"
                      }
                    },
                    0
                  ]
                },
                "else": true
              }
            },
            "message": "Wan3 720p Deal 不支持视频参考"
          },
          {
            "assert": {
              "$if": {
                "condition": {
                  "$and": [
                    {
                      "$eq": [
                        {
                          "$ref": "request.model"
                        },
                        "multiref:sd-2-c4"
                      ]
                    },
                    {
                      "$gt": [
                        {
                          "$len": {
                            "$filter": {
                              "from": {
                                "$sortByOrder": {
                                  "$ref": "request.images"
                                }
                              },
                              "as": "media",
                              "where": {
                                "$in": [
                                  {
                                    "$ref": "media.role"
                                  },
                                  [
                                    "first_frame",
                                    "last_frame"
                                  ]
                                ]
                              }
                            }
                          }
                        },
                        0
                      ]
                    }
                  ]
                },
                "then": {
                  "$and": [
                    {
                      "$eq": [
                        {
                          "$len": {
                            "$filter": {
                              "from": {
                                "$sortByOrder": {
                                  "$ref": "request.images"
                                }
                              },
                              "as": "media",
                              "where": {
                                "$in": [
                                  {
                                    "$ref": "media.role"
                                  },
                                  [
                                    "first_frame"
                                  ]
                                ]
                              }
                            }
                          }
                        },
                        1
                      ]
                    },
                    {
                      "$lte": [
                        {
                          "$len": {
                            "$filter": {
                              "from": {
                                "$sortByOrder": {
                                  "$ref": "request.images"
                                }
                              },
                              "as": "media",
                              "where": {
                                "$in": [
                                  {
                                    "$ref": "media.role"
                                  },
                                  [
                                    "last_frame"
                                  ]
                                ]
                              }
                            }
                          }
                        },
                        1
                      ]
                    },
                    {
                      "$eq": [
                        {
                          "$len": {
                            "$filter": {
                              "from": {
                                "$sortByOrder": {
                                  "$ref": "request.images"
                                }
                              },
                              "as": "media",
                              "where": {
                                "$in": [
                                  {
                                    "$ref": "media.role"
                                  },
                                  [
                                    "first_frame",
                                    "last_frame"
                                  ]
                                ]
                              }
                            }
                          }
                        },
                        {
                          "$len": {
                            "$ref": "request.images"
                          }
                        }
                      ]
                    },
                    {
                      "$eq": [
                        {
                          "$len": {
                            "$ref": "request.videos"
                          }
                        },
                        0
                      ]
                    },
                    {
                      "$eq": [
                        {
                          "$len": {
                            "$ref": "request.audios"
                          }
                        },
                        0
                      ]
                    }
                  ]
                },
                "else": true
              }
            },
            "message": "C4 首尾帧须有一张首帧、至多一张尾帧，不能混用参考图或音视频"
          }
        ],
        "lifecycle": {
          "requestBytes": 1048576,
          "createAttempts": 3,
          "pollIntervalMs": 5000,
          "anonymousResult": true,
          "mediaLimits": {
            "images": {
              "mimeTypes": [
                "image/png",
                "image/jpeg",
                "image/webp"
              ],
              "fileBytes": 10485760,
              "totalBytes": 104857600
            },
            "videos": {
              "mimeTypes": [
                "video/mp4",
                "video/quicktime",
                "video/webm"
              ],
              "fileBytes": 52428800,
              "totalBytes": 262144000
            },
            "audios": {
              "mimeTypes": [
                "audio/mpeg",
                "audio/wav",
                "audio/mp4",
                "audio/x-m4a",
                "audio/ogg",
                "audio/flac"
              ],
              "fileBytes": 52428800,
              "totalBytes": 262144000
            }
          },
          "prepare": [
            {
              "id": "catalog",
              "operation": {
                "method": "GET",
                "path": "/v1/catalog",
                "originPath": true
              },
              "expiresAt": {
                "$ref": "response.valid_until"
              },
              "value": {
                "revision": {
                  "$ref": "response.revision"
                },
                "model": {
                  "$first": {
                    "$filter": {
                      "from": {
                        "$ref": "response.models"
                      },
                      "as": "item",
                      "where": {
                        "$eq": [
                          {
                            "$ref": "item.id"
                          },
                          {
                            "$ref": "request.model"
                          }
                        ]
                      }
                    }
                  }
                }
              },
              "validations": [
                {
                  "assert": {
                    "$and": [
                      {
                        "$eq": [
                          {
                            "$ref": "response.status"
                          },
                          "ready"
                        ]
                      },
                      {
                        "$gt": [
                          {
                            "$len": {
                              "$ref": "response.revision"
                            }
                          },
                          0
                        ]
                      }
                    ]
                  },
                  "message": "上游目录尚未就绪，请稍后重试"
                },
                {
                  "assert": {
                    "$eq": [
                      {
                        "$len": {
                          "$filter": {
                            "from": {
                              "$ref": "response.models"
                            },
                            "as": "item",
                            "where": {
                              "$and": [
                                {
                                  "$eq": [
                                    {
                                      "$ref": "item.id"
                                    },
                                    {
                                      "$ref": "request.model"
                                    }
                                  ]
                                },
                                {
                                  "$eq": [
                                    {
                                      "$ref": "item.available"
                                    },
                                    true
                                  ]
                                }
                              ]
                            }
                          }
                        }
                      },
                      1
                    ]
                  },
                  "message": "所选模型当前不可用，未提交生成"
                },
                {
                  "assert": {
                    "$in": [
                      {
                        "$if": {
                          "condition": {
                            "$and": [
                              {
                                "$eq": [
                                  {
                                    "$ref": "request.model"
                                  },
                                  "multiref:sd-2-c4"
                                ]
                              },
                              {
                                "$gt": [
                                  {
                                    "$len": {
                                      "$filter": {
                                        "from": {
                                          "$sortByOrder": {
                                            "$ref": "request.images"
                                          }
                                        },
                                        "as": "media",
                                        "where": {
                                          "$in": [
                                            {
                                              "$ref": "media.role"
                                            },
                                            [
                                              "first_frame",
                                              "last_frame"
                                            ]
                                          ]
                                        }
                                      }
                                    }
                                  },
                                  0
                                ]
                              }
                            ]
                          },
                          "then": "frames",
                          "else": {
                            "$if": {
                              "condition": {
                                "$and": [
                                  {
                                    "$eq": [
                                      {
                                        "$ref": "request.model"
                                      },
                                      "qsqs:wan3-720p-deal"
                                    ]
                                  },
                                  {
                                    "$not": {
                                      "$gt": [
                                        {
                                          "$add": [
                                            {
                                              "$len": {
                                                "$ref": "request.images"
                                              }
                                            },
                                            {
                                              "$len": {
                                                "$ref": "request.videos"
                                              }
                                            },
                                            {
                                              "$len": {
                                                "$ref": "request.audios"
                                              }
                                            }
                                          ]
                                        },
                                        0
                                      ]
                                    }
                                  }
                                ]
                              },
                              "then": "text",
                              "else": "references"
                            }
                          }
                        }
                      },
                      {
                        "$get": [
                          {
                            "$first": {
                              "$filter": {
                                "from": {
                                  "$ref": "response.models"
                                },
                                "as": "item",
                                "where": {
                                  "$eq": [
                                    {
                                      "$ref": "item.id"
                                    },
                                    {
                                      "$ref": "request.model"
                                    }
                                  ]
                                }
                              }
                            }
                          },
                          "input_modes"
                        ]
                      }
                    ]
                  },
                  "message": "当前模型目录未声明所选输入模式"
                },
                {
                  "assert": {
                    "$if": {
                      "condition": {
                        "$gt": [
                          {
                            "$len": {
                              "$ref": "request.images"
                            }
                          },
                          0
                        ]
                      },
                      "then": {
                        "$gt": [
                          {
                            "$len": {
                              "$get": [
                                {
                                  "$get": [
                                    {
                                      "$first": {
                                        "$filter": {
                                          "from": {
                                            "$ref": "response.models"
                                          },
                                          "as": "item",
                                          "where": {
                                            "$eq": [
                                              {
                                                "$ref": "item.id"
                                              },
                                              {
                                                "$ref": "request.model"
                                              }
                                            ]
                                          }
                                        }
                                      }
                                    },
                                    "x-canvas-inputs"
                                  ]
                                },
                                "image"
                              ]
                            }
                          },
                          0
                        ]
                      },
                      "else": true
                    }
                  },
                  "message": "当前模型目录未声明 image 素材映射"
                },
                {
                  "assert": {
                    "$if": {
                      "condition": {
                        "$gt": [
                          {
                            "$len": {
                              "$ref": "request.videos"
                            }
                          },
                          0
                        ]
                      },
                      "then": {
                        "$gt": [
                          {
                            "$len": {
                              "$get": [
                                {
                                  "$get": [
                                    {
                                      "$first": {
                                        "$filter": {
                                          "from": {
                                            "$ref": "response.models"
                                          },
                                          "as": "item",
                                          "where": {
                                            "$eq": [
                                              {
                                                "$ref": "item.id"
                                              },
                                              {
                                                "$ref": "request.model"
                                              }
                                            ]
                                          }
                                        }
                                      }
                                    },
                                    "x-canvas-inputs"
                                  ]
                                },
                                "video"
                              ]
                            }
                          },
                          0
                        ]
                      },
                      "else": true
                    }
                  },
                  "message": "当前模型目录未声明 video 素材映射"
                },
                {
                  "assert": {
                    "$if": {
                      "condition": {
                        "$gt": [
                          {
                            "$len": {
                              "$ref": "request.audios"
                            }
                          },
                          0
                        ]
                      },
                      "then": {
                        "$gt": [
                          {
                            "$len": {
                              "$get": [
                                {
                                  "$get": [
                                    {
                                      "$first": {
                                        "$filter": {
                                          "from": {
                                            "$ref": "response.models"
                                          },
                                          "as": "item",
                                          "where": {
                                            "$eq": [
                                              {
                                                "$ref": "item.id"
                                              },
                                              {
                                                "$ref": "request.model"
                                              }
                                            ]
                                          }
                                        }
                                      }
                                    },
                                    "x-canvas-inputs"
                                  ]
                                },
                                "audio"
                              ]
                            }
                          },
                          0
                        ]
                      },
                      "else": true
                    }
                  },
                  "message": "当前模型目录未声明 audio 素材映射"
                }
              ]
            },
            {
              "id": "images",
              "source": "images",
              "operation": {
                "method": "POST",
                "path": "/v1/files",
                "contentType": "application/json",
                "body": {
                  "name": {
                    "$coalesce": [
                      {
                        "$omitEmpty": {
                          "$ref": "request.extra.prepareItem.name"
                        }
                      },
                      "reference-image"
                    ]
                  },
                  "image_b64": {
                    "$dataPayload": {
                      "$ref": "request.extra.prepareItem.dataUrl"
                    }
                  }
                },
                "originPath": true
              },
              "value": {
                "asset_id": {
                  "$ref": "response.asset_id"
                },
                "kind": "image",
                "role": {
                  "$ref": "request.extra.prepareItem.role"
                }
              },
              "validations": [
                {
                  "assert": {
                    "$and": [
                      {
                        "$gt": [
                          {
                            "$len": {
                              "$ref": "response.asset_id"
                            }
                          },
                          0
                        ]
                      },
                      {
                        "$eq": [
                          {
                            "$ref": "response.kind"
                          },
                          "image"
                        ]
                      }
                    ]
                  },
                  "message": "素材上传未返回对应类型的图库 asset_id"
                }
              ]
            },
            {
              "id": "videos",
              "source": "videos",
              "operation": {
                "method": "POST",
                "path": "/v1/files",
                "contentType": "application/json",
                "body": {
                  "name": {
                    "$coalesce": [
                      {
                        "$omitEmpty": {
                          "$ref": "request.extra.prepareItem.name"
                        }
                      },
                      "reference-video"
                    ]
                  },
                  "video_b64": {
                    "$dataPayload": {
                      "$ref": "request.extra.prepareItem.dataUrl"
                    }
                  }
                },
                "originPath": true
              },
              "value": {
                "asset_id": {
                  "$ref": "response.asset_id"
                },
                "kind": "video",
                "role": {
                  "$ref": "request.extra.prepareItem.role"
                }
              },
              "validations": [
                {
                  "assert": {
                    "$and": [
                      {
                        "$gt": [
                          {
                            "$len": {
                              "$ref": "response.asset_id"
                            }
                          },
                          0
                        ]
                      },
                      {
                        "$eq": [
                          {
                            "$ref": "response.kind"
                          },
                          "video"
                        ]
                      }
                    ]
                  },
                  "message": "素材上传未返回对应类型的图库 asset_id"
                }
              ]
            },
            {
              "id": "audios",
              "source": "audios",
              "operation": {
                "method": "POST",
                "path": "/v1/files",
                "contentType": "application/json",
                "body": {
                  "name": {
                    "$coalesce": [
                      {
                        "$omitEmpty": {
                          "$ref": "request.extra.prepareItem.name"
                        }
                      },
                      "reference-audio"
                    ]
                  },
                  "audio_b64": {
                    "$dataPayload": {
                      "$ref": "request.extra.prepareItem.dataUrl"
                    }
                  }
                },
                "originPath": true
              },
              "value": {
                "asset_id": {
                  "$ref": "response.asset_id"
                },
                "kind": "audio",
                "role": {
                  "$ref": "request.extra.prepareItem.role"
                }
              },
              "validations": [
                {
                  "assert": {
                    "$and": [
                      {
                        "$gt": [
                          {
                            "$len": {
                              "$ref": "response.asset_id"
                            }
                          },
                          0
                        ]
                      },
                      {
                        "$eq": [
                          {
                            "$ref": "response.kind"
                          },
                          "audio"
                        ]
                      }
                    ]
                  },
                  "message": "素材上传未返回对应类型的图库 asset_id"
                }
              ]
            },
            {
              "id": "quote",
              "operation": {
                "method": "POST",
                "path": "/v1/videos/quote",
                "contentType": "application/json",
                "body": {
                  "model_id": {
                    "$ref": "request.model"
                  },
                  "catalog_revision": {
                    "$ref": "request.extra.preparation.catalog.revision"
                  },
                  "input_mode": {
                    "$if": {
                      "condition": {
                        "$and": [
                          {
                            "$eq": [
                              {
                                "$ref": "request.model"
                              },
                              "multiref:sd-2-c4"
                            ]
                          },
                          {
                            "$gt": [
                              {
                                "$len": {
                                  "$filter": {
                                    "from": {
                                      "$sortByOrder": {
                                        "$ref": "request.images"
                                      }
                                    },
                                    "as": "media",
                                    "where": {
                                      "$in": [
                                        {
                                          "$ref": "media.role"
                                        },
                                        [
                                          "first_frame",
                                          "last_frame"
                                        ]
                                      ]
                                    }
                                  }
                                }
                              },
                              0
                            ]
                          }
                        ]
                      },
                      "then": "frames",
                      "else": {
                        "$if": {
                          "condition": {
                            "$and": [
                              {
                                "$eq": [
                                  {
                                    "$ref": "request.model"
                                  },
                                  "qsqs:wan3-720p-deal"
                                ]
                              },
                              {
                                "$not": {
                                  "$gt": [
                                    {
                                      "$add": [
                                        {
                                          "$len": {
                                            "$ref": "request.images"
                                          }
                                        },
                                        {
                                          "$len": {
                                            "$ref": "request.videos"
                                          }
                                        },
                                        {
                                          "$len": {
                                            "$ref": "request.audios"
                                          }
                                        }
                                      ]
                                    },
                                    0
                                  ]
                                }
                              }
                            ]
                          },
                          "then": "text",
                          "else": "references"
                        }
                      }
                    }
                  },
                  "input": {
                    "$merge": [
                      {
                        "prompt": {
                          "$ref": "request.prompt"
                        }
                      },
                      {
                        "$object": {
                          "from": {
                            "$unique": {
                              "$filter": {
                                "from": {
                                  "$map": {
                                    "from": [
                                      "image",
                                      "video",
                                      "audio"
                                    ],
                                    "as": "kind",
                                    "in": {
                                      "$get": [
                                        {
                                          "$merge": [
                                            {
                                              "$ref": "request.extra.preparation.catalog.model.x-canvas-inputs"
                                            },
                                            {
                                              "$if": {
                                                "condition": {
                                                  "$and": [
                                                    {
                                                      "$eq": [
                                                        {
                                                          "$ref": "request.model"
                                                        },
                                                        "multiref:sd-2-c4"
                                                      ]
                                                    },
                                                    {
                                                      "$gt": [
                                                        {
                                                          "$len": {
                                                            "$filter": {
                                                              "from": {
                                                                "$sortByOrder": {
                                                                  "$ref": "request.images"
                                                                }
                                                              },
                                                              "as": "media",
                                                              "where": {
                                                                "$in": [
                                                                  {
                                                                    "$ref": "media.role"
                                                                  },
                                                                  [
                                                                    "first_frame",
                                                                    "last_frame"
                                                                  ]
                                                                ]
                                                              }
                                                            }
                                                          }
                                                        },
                                                        0
                                                      ]
                                                    }
                                                  ]
                                                },
                                                "then": {
                                                  "image": "images"
                                                },
                                                "else": {}
                                              }
                                            }
                                          ]
                                        },
                                        {
                                          "$ref": "kind"
                                        }
                                      ]
                                    }
                                  }
                                },
                                "as": "field",
                                "where": {
                                  "$gt": [
                                    {
                                      "$len": {
                                        "$ref": "field"
                                      }
                                    },
                                    0
                                  ]
                                }
                              }
                            }
                          },
                          "as": "field",
                          "key": {
                            "$ref": "field"
                          },
                          "value": {
                            "$omitEmpty": {
                              "$map": {
                                "from": {
                                  "$filter": {
                                    "from": {
                                      "$concatArrays": [
                                        {
                                          "$if": {
                                            "condition": {
                                              "$and": [
                                                {
                                                  "$eq": [
                                                    {
                                                      "$ref": "request.model"
                                                    },
                                                    "multiref:sd-2-c4"
                                                  ]
                                                },
                                                {
                                                  "$gt": [
                                                    {
                                                      "$len": {
                                                        "$filter": {
                                                          "from": {
                                                            "$sortByOrder": {
                                                              "$ref": "request.images"
                                                            }
                                                          },
                                                          "as": "media",
                                                          "where": {
                                                            "$in": [
                                                              {
                                                                "$ref": "media.role"
                                                              },
                                                              [
                                                                "first_frame",
                                                                "last_frame"
                                                              ]
                                                            ]
                                                          }
                                                        }
                                                      }
                                                    },
                                                    0
                                                  ]
                                                }
                                              ]
                                            },
                                            "then": {
                                              "$concatArrays": [
                                                {
                                                  "$filter": {
                                                    "from": {
                                                      "$ref": "request.extra.preparation.images"
                                                    },
                                                    "as": "item",
                                                    "where": {
                                                      "$eq": [
                                                        {
                                                          "$ref": "item.role"
                                                        },
                                                        "first_frame"
                                                      ]
                                                    }
                                                  }
                                                },
                                                {
                                                  "$filter": {
                                                    "from": {
                                                      "$ref": "request.extra.preparation.images"
                                                    },
                                                    "as": "item",
                                                    "where": {
                                                      "$eq": [
                                                        {
                                                          "$ref": "item.role"
                                                        },
                                                        "last_frame"
                                                      ]
                                                    }
                                                  }
                                                }
                                              ]
                                            },
                                            "else": {
                                              "$ref": "request.extra.preparation.images"
                                            }
                                          }
                                        },
                                        {
                                          "$ref": "request.extra.preparation.videos"
                                        },
                                        {
                                          "$ref": "request.extra.preparation.audios"
                                        }
                                      ]
                                    },
                                    "as": "item",
                                    "where": {
                                      "$eq": [
                                        {
                                          "$get": [
                                            {
                                              "$merge": [
                                                {
                                                  "$ref": "request.extra.preparation.catalog.model.x-canvas-inputs"
                                                },
                                                {
                                                  "$if": {
                                                    "condition": {
                                                      "$and": [
                                                        {
                                                          "$eq": [
                                                            {
                                                              "$ref": "request.model"
                                                            },
                                                            "multiref:sd-2-c4"
                                                          ]
                                                        },
                                                        {
                                                          "$gt": [
                                                            {
                                                              "$len": {
                                                                "$filter": {
                                                                  "from": {
                                                                    "$sortByOrder": {
                                                                      "$ref": "request.images"
                                                                    }
                                                                  },
                                                                  "as": "media",
                                                                  "where": {
                                                                    "$in": [
                                                                      {
                                                                        "$ref": "media.role"
                                                                      },
                                                                      [
                                                                        "first_frame",
                                                                        "last_frame"
                                                                      ]
                                                                    ]
                                                                  }
                                                                }
                                                              }
                                                            },
                                                            0
                                                          ]
                                                        }
                                                      ]
                                                    },
                                                    "then": {
                                                      "image": "images"
                                                    },
                                                    "else": {}
                                                  }
                                                }
                                              ]
                                            },
                                            {
                                              "$ref": "item.kind"
                                            }
                                          ]
                                        },
                                        {
                                          "$ref": "field"
                                        }
                                      ]
                                    }
                                  }
                                },
                                "as": "item",
                                "in": {
                                  "asset_id": {
                                    "$ref": "item.asset_id"
                                  },
                                  "kind": {
                                    "$ref": "item.kind"
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                    ]
                  },
                  "parameters": {
                    "$object": {
                      "from": {
                        "$ref": "request.extra.preparation.catalog.model.parameter_schema.properties"
                      },
                      "as": "parameter",
                      "value": {
                        "$coalesce": [
                          {
                            "$switch": {
                              "cases": [
                                {
                                  "when": {
                                    "$eq": [
                                      {
                                        "$ref": "parameterKey"
                                      },
                                      "duration"
                                    ]
                                  },
                                  "then": {
                                    "$if": {
                                      "condition": {
                                        "$gt": [
                                          {
                                            "$ref": "request.duration"
                                          },
                                          0
                                        ]
                                      },
                                      "then": {
                                        "$ref": "request.duration"
                                      },
                                      "else": null
                                    }
                                  }
                                },
                                {
                                  "when": {
                                    "$eq": [
                                      {
                                        "$ref": "parameterKey"
                                      },
                                      "ratio"
                                    ]
                                  },
                                  "then": {
                                    "$omitEmpty": {
                                      "$ref": "request.aspectRatio"
                                    }
                                  }
                                },
                                {
                                  "when": {
                                    "$eq": [
                                      {
                                        "$ref": "parameterKey"
                                      },
                                      "resolution"
                                    ]
                                  },
                                  "then": {
                                    "$omitEmpty": {
                                      "$ref": "request.resolution"
                                    }
                                  }
                                },
                                {
                                  "when": {
                                    "$eq": [
                                      {
                                        "$ref": "parameterKey"
                                      },
                                      "quality"
                                    ]
                                  },
                                  "then": {
                                    "$coalesce": [
                                      {
                                        "$omitEmpty": {
                                          "$ref": "request.resolution"
                                        }
                                      },
                                      {
                                        "$omitEmpty": {
                                          "$ref": "request.quality"
                                        }
                                      }
                                    ]
                                  }
                                },
                                {
                                  "when": {
                                    "$eq": [
                                      {
                                        "$ref": "parameterKey"
                                      },
                                      "save_audio"
                                    ]
                                  },
                                  "then": {
                                    "$ref": "request.generateAudio"
                                  }
                                }
                              ],
                              "default": null
                            }
                          },
                          {
                            "$get": [
                              {
                                "$ref": "request.providerOptions.200t-vinted-mi.parameters"
                              },
                              {
                                "$ref": "parameterKey"
                              }
                            ]
                          },
                          {
                            "$ref": "parameter.default"
                          }
                        ]
                      }
                    }
                  }
                },
                "originPath": true
              },
              "value": {
                "body": {
                  "model_id": {
                    "$ref": "request.model"
                  },
                  "catalog_revision": {
                    "$ref": "request.extra.preparation.catalog.revision"
                  },
                  "input_mode": {
                    "$if": {
                      "condition": {
                        "$and": [
                          {
                            "$eq": [
                              {
                                "$ref": "request.model"
                              },
                              "multiref:sd-2-c4"
                            ]
                          },
                          {
                            "$gt": [
                              {
                                "$len": {
                                  "$filter": {
                                    "from": {
                                      "$sortByOrder": {
                                        "$ref": "request.images"
                                      }
                                    },
                                    "as": "media",
                                    "where": {
                                      "$in": [
                                        {
                                          "$ref": "media.role"
                                        },
                                        [
                                          "first_frame",
                                          "last_frame"
                                        ]
                                      ]
                                    }
                                  }
                                }
                              },
                              0
                            ]
                          }
                        ]
                      },
                      "then": "frames",
                      "else": {
                        "$if": {
                          "condition": {
                            "$and": [
                              {
                                "$eq": [
                                  {
                                    "$ref": "request.model"
                                  },
                                  "qsqs:wan3-720p-deal"
                                ]
                              },
                              {
                                "$not": {
                                  "$gt": [
                                    {
                                      "$add": [
                                        {
                                          "$len": {
                                            "$ref": "request.images"
                                          }
                                        },
                                        {
                                          "$len": {
                                            "$ref": "request.videos"
                                          }
                                        },
                                        {
                                          "$len": {
                                            "$ref": "request.audios"
                                          }
                                        }
                                      ]
                                    },
                                    0
                                  ]
                                }
                              }
                            ]
                          },
                          "then": "text",
                          "else": "references"
                        }
                      }
                    }
                  },
                  "input": {
                    "$merge": [
                      {
                        "prompt": {
                          "$ref": "request.prompt"
                        }
                      },
                      {
                        "$object": {
                          "from": {
                            "$unique": {
                              "$filter": {
                                "from": {
                                  "$map": {
                                    "from": [
                                      "image",
                                      "video",
                                      "audio"
                                    ],
                                    "as": "kind",
                                    "in": {
                                      "$get": [
                                        {
                                          "$merge": [
                                            {
                                              "$ref": "request.extra.preparation.catalog.model.x-canvas-inputs"
                                            },
                                            {
                                              "$if": {
                                                "condition": {
                                                  "$and": [
                                                    {
                                                      "$eq": [
                                                        {
                                                          "$ref": "request.model"
                                                        },
                                                        "multiref:sd-2-c4"
                                                      ]
                                                    },
                                                    {
                                                      "$gt": [
                                                        {
                                                          "$len": {
                                                            "$filter": {
                                                              "from": {
                                                                "$sortByOrder": {
                                                                  "$ref": "request.images"
                                                                }
                                                              },
                                                              "as": "media",
                                                              "where": {
                                                                "$in": [
                                                                  {
                                                                    "$ref": "media.role"
                                                                  },
                                                                  [
                                                                    "first_frame",
                                                                    "last_frame"
                                                                  ]
                                                                ]
                                                              }
                                                            }
                                                          }
                                                        },
                                                        0
                                                      ]
                                                    }
                                                  ]
                                                },
                                                "then": {
                                                  "image": "images"
                                                },
                                                "else": {}
                                              }
                                            }
                                          ]
                                        },
                                        {
                                          "$ref": "kind"
                                        }
                                      ]
                                    }
                                  }
                                },
                                "as": "field",
                                "where": {
                                  "$gt": [
                                    {
                                      "$len": {
                                        "$ref": "field"
                                      }
                                    },
                                    0
                                  ]
                                }
                              }
                            }
                          },
                          "as": "field",
                          "key": {
                            "$ref": "field"
                          },
                          "value": {
                            "$omitEmpty": {
                              "$map": {
                                "from": {
                                  "$filter": {
                                    "from": {
                                      "$concatArrays": [
                                        {
                                          "$if": {
                                            "condition": {
                                              "$and": [
                                                {
                                                  "$eq": [
                                                    {
                                                      "$ref": "request.model"
                                                    },
                                                    "multiref:sd-2-c4"
                                                  ]
                                                },
                                                {
                                                  "$gt": [
                                                    {
                                                      "$len": {
                                                        "$filter": {
                                                          "from": {
                                                            "$sortByOrder": {
                                                              "$ref": "request.images"
                                                            }
                                                          },
                                                          "as": "media",
                                                          "where": {
                                                            "$in": [
                                                              {
                                                                "$ref": "media.role"
                                                              },
                                                              [
                                                                "first_frame",
                                                                "last_frame"
                                                              ]
                                                            ]
                                                          }
                                                        }
                                                      }
                                                    },
                                                    0
                                                  ]
                                                }
                                              ]
                                            },
                                            "then": {
                                              "$concatArrays": [
                                                {
                                                  "$filter": {
                                                    "from": {
                                                      "$ref": "request.extra.preparation.images"
                                                    },
                                                    "as": "item",
                                                    "where": {
                                                      "$eq": [
                                                        {
                                                          "$ref": "item.role"
                                                        },
                                                        "first_frame"
                                                      ]
                                                    }
                                                  }
                                                },
                                                {
                                                  "$filter": {
                                                    "from": {
                                                      "$ref": "request.extra.preparation.images"
                                                    },
                                                    "as": "item",
                                                    "where": {
                                                      "$eq": [
                                                        {
                                                          "$ref": "item.role"
                                                        },
                                                        "last_frame"
                                                      ]
                                                    }
                                                  }
                                                }
                                              ]
                                            },
                                            "else": {
                                              "$ref": "request.extra.preparation.images"
                                            }
                                          }
                                        },
                                        {
                                          "$ref": "request.extra.preparation.videos"
                                        },
                                        {
                                          "$ref": "request.extra.preparation.audios"
                                        }
                                      ]
                                    },
                                    "as": "item",
                                    "where": {
                                      "$eq": [
                                        {
                                          "$get": [
                                            {
                                              "$merge": [
                                                {
                                                  "$ref": "request.extra.preparation.catalog.model.x-canvas-inputs"
                                                },
                                                {
                                                  "$if": {
                                                    "condition": {
                                                      "$and": [
                                                        {
                                                          "$eq": [
                                                            {
                                                              "$ref": "request.model"
                                                            },
                                                            "multiref:sd-2-c4"
                                                          ]
                                                        },
                                                        {
                                                          "$gt": [
                                                            {
                                                              "$len": {
                                                                "$filter": {
                                                                  "from": {
                                                                    "$sortByOrder": {
                                                                      "$ref": "request.images"
                                                                    }
                                                                  },
                                                                  "as": "media",
                                                                  "where": {
                                                                    "$in": [
                                                                      {
                                                                        "$ref": "media.role"
                                                                      },
                                                                      [
                                                                        "first_frame",
                                                                        "last_frame"
                                                                      ]
                                                                    ]
                                                                  }
                                                                }
                                                              }
                                                            },
                                                            0
                                                          ]
                                                        }
                                                      ]
                                                    },
                                                    "then": {
                                                      "image": "images"
                                                    },
                                                    "else": {}
                                                  }
                                                }
                                              ]
                                            },
                                            {
                                              "$ref": "item.kind"
                                            }
                                          ]
                                        },
                                        {
                                          "$ref": "field"
                                        }
                                      ]
                                    }
                                  }
                                },
                                "as": "item",
                                "in": {
                                  "asset_id": {
                                    "$ref": "item.asset_id"
                                  },
                                  "kind": {
                                    "$ref": "item.kind"
                                  }
                                }
                              }
                            }
                          }
                        }
                      }
                    ]
                  },
                  "parameters": {
                    "$object": {
                      "from": {
                        "$ref": "request.extra.preparation.catalog.model.parameter_schema.properties"
                      },
                      "as": "parameter",
                      "value": {
                        "$coalesce": [
                          {
                            "$switch": {
                              "cases": [
                                {
                                  "when": {
                                    "$eq": [
                                      {
                                        "$ref": "parameterKey"
                                      },
                                      "duration"
                                    ]
                                  },
                                  "then": {
                                    "$if": {
                                      "condition": {
                                        "$gt": [
                                          {
                                            "$ref": "request.duration"
                                          },
                                          0
                                        ]
                                      },
                                      "then": {
                                        "$ref": "request.duration"
                                      },
                                      "else": null
                                    }
                                  }
                                },
                                {
                                  "when": {
                                    "$eq": [
                                      {
                                        "$ref": "parameterKey"
                                      },
                                      "ratio"
                                    ]
                                  },
                                  "then": {
                                    "$omitEmpty": {
                                      "$ref": "request.aspectRatio"
                                    }
                                  }
                                },
                                {
                                  "when": {
                                    "$eq": [
                                      {
                                        "$ref": "parameterKey"
                                      },
                                      "resolution"
                                    ]
                                  },
                                  "then": {
                                    "$omitEmpty": {
                                      "$ref": "request.resolution"
                                    }
                                  }
                                },
                                {
                                  "when": {
                                    "$eq": [
                                      {
                                        "$ref": "parameterKey"
                                      },
                                      "quality"
                                    ]
                                  },
                                  "then": {
                                    "$coalesce": [
                                      {
                                        "$omitEmpty": {
                                          "$ref": "request.resolution"
                                        }
                                      },
                                      {
                                        "$omitEmpty": {
                                          "$ref": "request.quality"
                                        }
                                      }
                                    ]
                                  }
                                },
                                {
                                  "when": {
                                    "$eq": [
                                      {
                                        "$ref": "parameterKey"
                                      },
                                      "save_audio"
                                    ]
                                  },
                                  "then": {
                                    "$ref": "request.generateAudio"
                                  }
                                }
                              ],
                              "default": null
                            }
                          },
                          {
                            "$get": [
                              {
                                "$ref": "request.providerOptions.200t-vinted-mi.parameters"
                              },
                              {
                                "$ref": "parameterKey"
                              }
                            ]
                          },
                          {
                            "$ref": "parameter.default"
                          }
                        ]
                      }
                    }
                  }
                },
                "id": {
                  "$ref": "response.quote_id"
                }
              },
              "validations": [
                {
                  "assert": {
                    "$and": [
                      {
                        "$gt": [
                          {
                            "$len": {
                              "$ref": "response.quote_id"
                            }
                          },
                          0
                        ]
                      },
                      {
                        "$ne": [
                          {
                            "$ref": "response.credits"
                          },
                          null
                        ]
                      },
                      {
                        "$eq": [
                          {
                            "$ref": "response.credits"
                          },
                          {
                            "$toInt": {
                              "$ref": "response.credits"
                            }
                          }
                        ]
                      },
                      {
                        "$gte": [
                          {
                            "$ref": "response.credits"
                          },
                          0
                        ]
                      },
                      {
                        "$eq": [
                          {
                            "$ref": "response.catalog_revision"
                          },
                          {
                            "$ref": "request.extra.preparation.catalog.revision"
                          }
                        ]
                      }
                    ]
                  },
                  "message": "报价缺少有效编号、整数积分或目录版本，未提交生成"
                }
              ],
              "message": {
                "$concat": [
                  "200t.vinted.mi 本次上游报价：",
                  {
                    "$ref": "response.credits"
                  },
                  " 积分；操作编号：",
                  {
                    "$ref": "request.extra.idempotencyKey"
                  }
                ]
              }
            }
          ]
        },
        "create": {
          "method": "POST",
          "path": "/v1/videos",
          "contentType": "application/json",
          "body": {
            "$ref": "request.extra.preparation.quote.body"
          },
          "originPath": true,
          "headers": {
            "Idempotency-Key": {
              "$ref": "request.extra.idempotencyKey"
            },
            "Expected-Quote-Id": {
              "$ref": "request.extra.preparation.quote.id"
            }
          }
        },
        "poll": {
          "method": "GET",
          "path": "/v1/videos/{{taskId}}",
          "originPath": true
        },
        "result": {
          "method": "GET",
          "path": "/v1/videos/{{taskId}}/signed_url",
          "originPath": true
        },
        "response": {
          "taskId": {
            "$coalesce": [
              {
                "$ref": "taskId"
              },
              {
                "$ref": "response.id"
              }
            ]
          },
          "status": {
            "$if": {
              "condition": {
                "$and": [
                  {
                    "$ref": "taskId"
                  },
                  {
                    "$ref": "response.id"
                  },
                  {
                    "$ne": [
                      {
                        "$ref": "response.id"
                      },
                      {
                        "$ref": "taskId"
                      }
                    ]
                  }
                ]
              },
              "then": "failed",
              "else": {
                "$if": {
                  "condition": {
                    "$and": [
                      {
                        "$not": {
                          "$ref": "response.status"
                        }
                      },
                      {
                        "$coalesce": [
                          {
                            "$ref": "response.url"
                          },
                          {
                            "$ref": "response.download_url"
                          }
                        ]
                      }
                    ]
                  },
                  "then": "succeeded",
                  "else": {
                    "$if": {
                      "condition": {
                        "$and": [
                          {
                            "$or": [
                              {
                                "$eq": [
                                  {
                                    "$ref": "response.status"
                                  },
                                  "succeeded"
                                ]
                              },
                              {
                                "$and": [
                                  {
                                    "$eq": [
                                      {
                                        "$ref": "response.status"
                                      },
                                      "failed"
                                    ]
                                  },
                                  {
                                    "$eq": [
                                      {
                                        "$ref": "response.refunded"
                                      },
                                      true
                                    ]
                                  },
                                  {
                                    "$eq": [
                                      {
                                        "$ref": "response.has_file"
                                      },
                                      true
                                    ]
                                  },
                                  {
                                    "$eq": [
                                      {
                                        "$ref": "response.recovered_artifact.available"
                                      },
                                      true
                                    ]
                                  },
                                  {
                                    "$eq": [
                                      {
                                        "$ref": "response.recovered_artifact.settlement_unchanged"
                                      },
                                      true
                                    ]
                                  }
                                ]
                              }
                            ]
                          },
                          {
                            "$eq": [
                              {
                                "$ref": "response.has_file"
                              },
                              true
                            ]
                          }
                        ]
                      },
                      "then": "succeeded",
                      "else": {
                        "$if": {
                          "condition": {
                            "$eq": [
                              {
                                "$ref": "response.status"
                              },
                              "failed"
                            ]
                          },
                          "then": "failed",
                          "else": {
                            "$if": {
                              "condition": {
                                "$eq": [
                                  {
                                    "$ref": "response.status"
                                  },
                                  "queued"
                                ]
                              },
                              "then": "pending",
                              "else": "processing"
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          },
          "videos": {
            "$if": {
              "condition": {
                "$not": {
                  "$ref": "response.status"
                }
              },
              "then": {
                "$coalesce": [
                  {
                    "$ref": "response.url"
                  },
                  {
                    "$ref": "response.download_url"
                  }
                ]
              },
              "else": null
            }
          },
          "message": {
            "$coalesce": [
              {
                "$ref": "response.error.message"
              },
              {
                "$ref": "response.error"
              },
              {
                "$ref": "response.progress.message"
              }
            ]
          },
          "usage": {
            "$ref": "response.result.actual"
          },
          "resultEphemeral": true
        }
      }
    ]
  }
}
```
<!-- YINGCE_MANIFEST_CONTRACT_END -->
