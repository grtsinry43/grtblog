import CategoryController from '@/services/category/CategoryController';
import { getToken } from '@/utils/token';
import { useDispatch, useSelector } from '@@/exports';
import {
  CopyrightOutlined,
  FileMarkdownOutlined,
  FireOutlined,
  PlusOutlined,
  PushpinOutlined,
  QuestionCircleOutlined,
  TagOutlined,
} from '@ant-design/icons';
import { PageContainer, ProCard, ProForm } from '@ant-design/pro-components';
import {
  Button,
  Card,
  Col,
  Divider,
  Form,
  Image,
  Input,
  message,
  Modal,
  Row,
  Select,
  Space,
  Switch,
  Tooltip,
  Typography,
  Upload,
} from 'antd';
import { MdEditor } from 'md-editor-rt';
import 'md-editor-rt/lib/style.css';
import React, { useEffect, useState } from 'react';

// 导入额外的插件
import highlight from 'highlight.js';
import 'highlight.js/styles/atom-one-dark.css';
import Katex from 'katex';
import 'katex/dist/katex.min.css';
import mermaid from 'mermaid';

const { Title, Text } = Typography;
const { Option } = Select;

interface ArticleFormProps {
  type: 'add' | 'edit';
  articleInfo: any;
  setArticleInfo: (info: any) => void;
  submitHandle: (content: string) => void;
}

const ArticleForm: React.FC<ArticleFormProps> = ({
  type,
  articleInfo,
  setArticleInfo,
  submitHandle,
}) => {
  const [form] = Form.useForm();
  const { list } = useSelector((state: any) => state.category);
  const dispatch = useDispatch();
  const [editorContent, setEditorContent] = useState('');
  const [firstIn, setFirstIn] = useState(true);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [addCategoryForm, setAddCategoryForm] = useState({
    name: '',
    shortUrl: '',
    type: 1,
  });
  const [editorId] = useState(
    'md-editor-rt-' + Math.random().toString(36).substr(2, 9),
  );

  // 初始化 mermaid
  useEffect(() => {
    mermaid.initialize({
      startOnLoad: true,
      theme: 'default',
      securityLevel: 'loose',
    });
  }, []);

  // 初始化分类列表
  useEffect(() => {
    if (list.length === 0) {
      dispatch({
        type: 'category/initCategoryList',
      });
    }
  }, []);

  // 表单初始化
  useEffect(() => {
    if (form && articleInfo && firstIn) {
      form.setFieldsValue(articleInfo);
      setEditorContent(articleInfo?.content || '');
      setFirstIn(false);
    }
  }, [articleInfo, firstIn, form]);

  const onValueChange = (key: string, value: any) => {
    setArticleInfo({
      ...articleInfo,
      [key]: value,
    });
  };

  const handleSubmit = () => {
    submitHandle(editorContent);
  };

  const handleAddCategory = () => {
    if (!addCategoryForm.name) {
      message.error('请输入分类名称');
      return;
    }
    if (!addCategoryForm.shortUrl) {
      message.error('请输入分类短链接');
      return;
    }

    CategoryController.addCategoryApi(addCategoryForm).then((res) => {
      if (res) {
        message.success('分类添加成功');
        setIsModalVisible(false);
        setAddCategoryForm({
          name: '',
          shortUrl: '',
          type: 1,
        });
        dispatch({
          type: 'category/initCategoryList',
        });
      } else {
        message.error('分类添加失败');
      }
    });
  };

  // 上传图片到编辑器
  const onUploadImg = async (
    files: File[],
    callback: (urls: string[]) => void,
  ) => {
    const res: string[] = [];

    for (const file of files) {
      const formData = new FormData();
      formData.append('file', file);

      try {
        const response = await fetch('/api/v1/upload', {
          method: 'POST',
          headers: {
            Authorization: 'Bearer ' + getToken(),
          },
          body: formData,
        });

        const data = await response.json();
        if (data && data.data) {
          res.push(data.data);
        }
      } catch (error) {
        console.error('Upload failed:', error);
      }
    }

    callback(res);
  };

  // 处理 Markdown 文件导入
  const handleMarkdownImport = async (file: File) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      const content = e.target?.result as string;
      setEditorContent(content);
    };
    reader.readAsText(file);
    return false; // 阻止默认上传行为
  };

  return (
    <PageContainer
      title={type === 'add' ? '添加文章' : '编辑文章'}
      subTitle={
        type === 'add'
          ? '今天想写点什么呢 ٩(๑˃̵ᴗ˂̵๑)۶'
          : '每一次的雕琢都是成就完美的作品哇'
      }
    >
      <ProCard>
        <ProForm
          form={form}
          onFinish={handleSubmit}
          submitter={{
            render: (props) => {
              return (
                <Row justify="center" style={{ marginTop: 24 }}>
                  <Button
                    type="primary"
                    size="large"
                    onClick={() => props.form?.submit()}
                  >
                    {type === 'add' ? '发布文章' : '保存修改'}
                  </Button>
                </Row>
              );
            },
          }}
        >
          <Row gutter={24}>
            <Col span={16}>
              <Card
                title={
                  <Space>
                    <span>文章内容</span>
                    <Tooltip title="支持 Markdown 语法、数学公式、流程图、代码高亮等功能">
                      <QuestionCircleOutlined />
                    </Tooltip>
                  </Space>
                }
                bordered={false}
              >
                <Form.Item
                  label="标题"
                  name="title"
                  rules={[{ required: true, message: '请输入标题' }]}
                >
                  <Input
                    size="large"
                    placeholder="请输入文章标题"
                    onChange={(e) => onValueChange('title', e.target.value)}
                  />
                </Form.Item>

                <Form.Item label="内容编辑" name="content">
                  <div
                    style={{ border: '1px solid #d9d9d9', borderRadius: '2px' }}
                  >
                    <MdEditor
                      modelValue={editorContent}
                      onChange={setEditorContent}
                      id={editorId}
                      style={{ height: '600px' }}
                      // 工具栏配置
                      toolbars={[
                        'bold',
                        'italic',
                        // 'strikethrough',
                        'sub',
                        'sup',
                        'quote',
                        'unorderedList',
                        'orderedList',
                        'codeRow',
                        'code',
                        'link',
                        'image',
                        'table',
                        'mermaid',
                        'katex',
                        'revoke',
                        'next',
                        'save',
                        'pageFullscreen',
                        'fullscreen',
                        'preview',
                        'htmlPreview',
                        'catalog',
                        'github',
                      ]}
                      // 扩展配置
                      onUploadImg={onUploadImg}
                      // 代码高亮
                        // @ts-ignore
                      codeHighlightExtensionMap={{
                        vue: highlight.getLanguage('vue'),
                        typescript: highlight.getLanguage('typescript'),
                        javascript: highlight.getLanguage('javascript'),
                        css: highlight.getLanguage('css'),
                        html: highlight.getLanguage('html'),
                        go: highlight.getLanguage('go'),
                        java: highlight.getLanguage('java'),
                        python: highlight.getLanguage('python'),
                        rust: highlight.getLanguage('rust'),
                      }}
                      // 数学公式
                      showCodeRowNumber={true}
                      // 预览主题
                      previewTheme="vuepress"
                      // 扩展配置
                      // extensions={[
                      //   {
                      //     name: 'mermaid',
                      //     level: 'block',
                      //     tokenizer: (source) => {
                      //       const match = source.match(
                      //         /^```mermaid\s+([\s\S]+?)```/,
                      //       );
                      //       if (match) {
                      //         return {
                      //           type: 'mermaid',
                      //           raw: match[0],
                      //           content: match[1].trim(),
                      //           tokens: [],
                      //         };
                      //       }
                      //       return undefined;
                      //     },
                      //     renderer: (token) => {
                      //       return `<div class="mermaid">${token.content}</div>`;
                      //     },
                      //   },
                      //   {
                      //     name: 'katex',
                      //     level: 'inline',
                      //     tokenizer: (source) => {
                      //       const match = source.match(/^\$\$(.*?)\$\$/);
                      //       if (match) {
                      //         return {
                      //           type: 'katex',
                      //           raw: match[0],
                      //           content: match[1].trim(),
                      //           tokens: [],
                      //         };
                      //       }
                      //       return undefined;
                      //     },
                      //     renderer: (token) => {
                      //       try {
                      //         return `<span class="katex">${Katex.renderToString(
                      //           token.content,
                      //           { throwOnError: false },
                      //         )}</span>`;
                      //       } catch (e) {
                      //         return `<span class="katex-error">${token.content}</span>`;
                      //       }
                      //     },
                      //   },
                      // ]}
                    />
                  </div>
                </Form.Item>

                <Divider dashed />

                <Space direction="vertical" style={{ width: '100%' }}>
                  <Text type="secondary">从 Markdown 文件导入内容</Text>
                  <Upload
                    accept=".md,.markdown"
                    showUploadList={false}
                    beforeUpload={handleMarkdownImport}
                    maxCount={1}
                  >
                    <Button icon={<FileMarkdownOutlined />}>
                      选择 Markdown 文件
                    </Button>
                  </Upload>
                </Space>
              </Card>
            </Col>

            <Col span={8}>
              <Card title="文章设置" bordered={false}>
                <Form.Item
                  label="分类"
                  name="categoryId"
                  // rules={[{ required: true, message: '请选择分类' }]}
                >
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Select
                      showSearch
                      placeholder="选择文章分类"
                      optionFilterProp="children"
                      onChange={(value) => onValueChange('categoryId', value)}
                      style={{ width: '100%' }}
                    >
                      {list.map((item: any) => (
                        <Option key={item.id} value={item.id}>
                          {item.name}
                        </Option>
                      ))}
                    </Select>
                    <Button type="link" onClick={() => setIsModalVisible(true)}>
                      + 创建新分类
                    </Button>
                  </Space>
                </Form.Item>

                <Form.Item
                  label="短链接"
                  name="shortUrl"
                  tooltip="自定义文章链接，留空则根据标题自动生成"
                >
                  <Input
                    placeholder="例如：hello-world"
                    onChange={(e) => onValueChange('shortUrl', e.target.value)}
                  />
                </Form.Item>

                <Form.Item
                  label="标签"
                  name="tags"
                  rules={[{ required: true, message: '请输入标签' }]}
                >
                  <Input
                    prefix={<TagOutlined />}
                    placeholder="使用英文逗号分隔，如：技术,教程,React"
                    onChange={(e) => onValueChange('tags', e.target.value)}
                  />
                </Form.Item>

                <Form.Item label="文章封面" name="cover">
                  {type === 'edit' && articleInfo?.cover && (
                    <Image
                      src={articleInfo.cover || '/placeholder.svg'}
                      width={200}
                      style={{ marginBottom: '16px' }}
                    />
                  )}
                  <Upload
                    listType="picture-card"
                    maxCount={1}
                    action="/api/v1/upload"
                    headers={{
                      Authorization: 'Bearer ' + getToken(),
                    }}
                    onChange={(e) => {
                      if (e.file.status === 'done') {
                        const url = e.file.response.data;
                        onValueChange('cover', url);
                      }
                    }}
                  >
                    <div>
                      <PlusOutlined />
                      <div style={{ marginTop: 8 }}>上传封面</div>
                    </div>
                  </Upload>
                </Form.Item>

                <Form.Item
                  label="发布状态"
                  name="isPublished"
                  rules={[{ required: true, message: '请选择文章状态' }]}
                >
                  <Select
                    placeholder="选择文章状态"
                    onChange={(value) => onValueChange('isPublished', value)}
                  >
                    <Option value={false}>保存为草稿</Option>
                    <Option value={true}>立即发布</Option>
                  </Select>
                </Form.Item>

                <Divider orientation="left">文章属性</Divider>

                <Row gutter={[16, 16]}>
                  <Col span={8}>
                    <Form.Item name="isTop" valuePropName="checked">
                      <div style={{ textAlign: 'center' }}>
                        <Switch
                          checkedChildren={<PushpinOutlined />}
                          unCheckedChildren={<PushpinOutlined />}
                          checked={articleInfo?.isTop}
                          onChange={(checked) =>
                            onValueChange('isTop', checked)
                          }
                        />
                        <div style={{ marginTop: 8 }}>置顶</div>
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={8}>
                    <Form.Item name="isHot" valuePropName="checked">
                      <div style={{ textAlign: 'center' }}>
                        <Switch
                          checkedChildren={<FireOutlined />}
                          unCheckedChildren={<FireOutlined />}
                          checked={articleInfo?.isHot}
                          onChange={(checked) =>
                            onValueChange('isHot', checked)
                          }
                        />
                        <div style={{ marginTop: 8 }}>热门</div>
                      </div>
                    </Form.Item>
                  </Col>
                  <Col span={8}>
                    <Form.Item name="isOriginal" valuePropName="checked">
                      <div style={{ textAlign: 'center' }}>
                        <Switch
                          checkedChildren={<CopyrightOutlined />}
                          unCheckedChildren={<CopyrightOutlined />}
                          checked={articleInfo?.isOriginal}
                          onChange={(checked) =>
                            onValueChange('isOriginal', checked)
                          }
                        />
                        <div style={{ marginTop: 8 }}>原创</div>
                      </div>
                    </Form.Item>
                  </Col>
                </Row>

                <Divider />

                <Card
                  title="Markdown 编辑器功能"
                  size="small"
                  bordered={false}
                  style={{ marginBottom: '16px' }}
                >
                  <ul style={{ paddingLeft: '20px', margin: 0 }}>
                    <li>支持标准 Markdown 语法</li>
                    <li>数学公式（KaTeX）：$$E=mc^2$$</li>
                    <li>流程图（Mermaid）</li>
                    <li>代码高亮（多种语言）</li>
                    <li>表格、列表、引用</li>
                    <li>图片上传与管理</li>
                    <li>全屏编辑与预览</li>
                  </ul>
                </Card>
              </Card>
            </Col>
          </Row>
        </ProForm>
      </ProCard>

      {/* 新建分类弹窗 */}
      <Modal
        title="新建分类"
        open={isModalVisible}
        onOk={handleAddCategory}
        onCancel={() => setIsModalVisible(false)}
      >
        <Form layout="vertical">
          <Form.Item
            label="分类名称"
            required
            rules={[{ required: true, message: '请输入分类名称' }]}
          >
            <Input
              placeholder="请输入分类名称"
              value={addCategoryForm.name}
              onChange={(e) =>
                setAddCategoryForm({
                  ...addCategoryForm,
                  name: e.target.value,
                })
              }
            />
          </Form.Item>
          <Form.Item
            label="分类短链接"
            required
            rules={[{ required: true, message: '请输入分类短链接' }]}
          >
            <Input
              placeholder="请输入分类短链接，如：tech-blog"
              value={addCategoryForm.shortUrl}
              onChange={(e) =>
                setAddCategoryForm({
                  ...addCategoryForm,
                  shortUrl: e.target.value,
                })
              }
            />
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default ArticleForm;
