'use client';

import useRouteLeaveConfirm from '@/hooks/use-route-leave-confirm';
import { getAllConfig, updateConfig } from '@/services/config/ConfigController';
import { refreshFrontendCache } from '@/services/refersh';
import {
  CheckCircleOutlined,
  CopyrightOutlined,
  EditOutlined,
  FieldTimeOutlined,
  GithubOutlined,
  GlobalOutlined,
  HomeOutlined,
  LinkOutlined,
  PictureOutlined,
  ReloadOutlined,
  SafetyCertificateOutlined,
  SaveOutlined,
  TagsOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import {
  Avatar,
  Button,
  Card,
  Col,
  Divider,
  Empty,
  Form,
  Input,
  message,
  Row,
  Space,
  Spin,
  Tabs,
  Tooltip,
  Typography,
} from 'antd';
import React, { useEffect, useState } from 'react';

const { Title, Text, Paragraph } = Typography;
const { TabPane } = Tabs;

const ConfigPage = () => {
  const [config, setConfig] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [activeTab, setActiveTab] = useState('1');
  const [formChanged, setFormChanged] = useState(false);
  const formRef = React.useRef<any>(null);
  useRouteLeaveConfirm();

  const fetchConfig = () => {
    setLoading(true);
    getAllConfig()
      .then((res) => {
        if (res && res.data) {
          formRef.current?.setFieldsValue(res.data);
          setConfig(res.data);
        }
      })
      .catch((e) => {
        console.log(e);
        message.error('获取配置失败');
      })
      .finally(() => {
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchConfig();
  }, []);

  const onChangeHandle = (name: string, value: any) => {
    setConfig({
      ...config,
      [name]: value,
    });
    setFormChanged(true);
  };

  const handleUpdate = (values: any) => {
    setSubmitting(true);
    const updatePromises = Object.keys(values).map((key) => {
      return updateConfig({ key, value: values[key] });
    });

    Promise.all(updatePromises)
      .then(() => {
        message.success({
          content: '配置更新成功',
          icon: <CheckCircleOutlined style={{ color: '#52c41a' }} />,
        });
        refreshFrontendCache().then((res) => {
          if (res) {
            message.success({
              content: '刷新缓存成功',
              icon: <CheckCircleOutlined style={{ color: '#52c41a' }} />,
            });
          } else {
            message.error('刷新缓存失败');
          }
        });
        setFormChanged(false);
      })
      .catch(() => {
        message.error('配置更新失败');
      })
      .finally(() => {
        setSubmitting(false);
      });
  };

  const renderPreviewSection = () => {
    return (
      <div className="preview-section">
        <Divider orientation="left">
          <Space>
            <PictureOutlined />
            <span>预览效果</span>
          </Space>
        </Divider>

        <Row gutter={[24, 24]} justify="center" align="middle">
          <Col xs={24} md={12}>
            <Card bordered={false} className="preview-card">
              <div style={{ textAlign: 'center' }}>
                <Avatar
                  src={config.AUTHOR_AVATAR || '/placeholder-avatar.png'}
                  size={100}
                  style={{ border: '4px solid #f0f0f0' }}
                />
                <Title level={3} style={{ marginTop: 16, marginBottom: 4 }}>
                  {config.AUTHOR_NAME || '作者姓名'}
                </Title>
                <Paragraph type="secondary" style={{ marginBottom: 16 }}>
                  {config.AUTHOR_INFO || '作者简介信息将显示在这里'}
                </Paragraph>
                <Space size="middle">
                  {config.AUTHOR_GITHUB && (
                    <Tooltip title="GitHub">
                      <a
                        href={config.AUTHOR_GITHUB}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        <Button shape="circle" icon={<GithubOutlined />} />
                      </a>
                    </Tooltip>
                  )}
                  {config.AUTHOR_HOME && (
                    <Tooltip title="个人主页">
                      <a
                        href={config.AUTHOR_HOME}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        <Button shape="circle" icon={<HomeOutlined />} />
                      </a>
                    </Tooltip>
                  )}
                </Space>
                <Divider dashed />
                <Paragraph>
                  <Text italic>
                    &#34;{config.AUTHOR_WELCOME || '欢迎信息将显示在这里'}&#34;
                  </Text>
                </Paragraph>
              </div>
            </Card>
          </Col>

          <Col xs={24} md={12}>
            <Card bordered={false} className="preview-card">
              <div style={{ textAlign: 'center' }}>
                {config.WEBSITE_LOGO && (
                  <img
                    src={config.WEBSITE_LOGO || '/placeholder.svg'}
                    alt="网站Logo"
                    style={{
                      maxHeight: 60,
                      maxWidth: '100%',
                      marginBottom: 16,
                    }}
                  />
                )}
                <Title level={3} style={{ marginBottom: 4 }}>
                  {config.WEBSITE_NAME || '网站名称'}
                </Title>
                <Paragraph type="secondary" style={{ marginBottom: 16 }}>
                  {config.WEBSITE_DESCRIPTION || '网站描述信息将显示在这里'}
                </Paragraph>
                <div>
                  <Title level={4} style={{ marginBottom: 8 }}>
                    {config.HOME_TITLE || '主页标题'}
                  </Title>
                  <Paragraph strong>
                    {config.HOME_SLOGAN || '主页标语'}
                  </Paragraph>
                  <Paragraph italic type="secondary">
                    {config.HOME_SLOGAN_EN || 'English Slogan'}
                  </Paragraph>
                </div>
              </div>
            </Card>
          </Col>
        </Row>
      </div>
    );
  };

  if (loading) {
    return (
      <PageContainer title="网站配置">
        <div className="loading-container">
          <Spin size="large" tip="加载配置中..." />
        </div>
      </PageContainer>
    );
  }

  return (
    <PageContainer
      title="网站配置"
      subTitle="管理网站的基本信息和显示内容"
      extra={[
        <Button key="refresh" icon={<ReloadOutlined />} onClick={fetchConfig}>
          刷新
        </Button>,
        <Button
          key="save"
          type="primary"
          icon={<SaveOutlined />}
          onClick={() => formRef.current.submit()}
          loading={submitting}
          disabled={!formChanged}
        >
          保存配置
        </Button>,
      ]}
    >
      {renderPreviewSection()}

      <Form
        onFinish={handleUpdate}
        ref={formRef}
        layout="vertical"
        initialValues={config}
        disabled={loading}
        className="config-form"
      >
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          type="card"
          className="config-tabs"
        >
          <TabPane
            tab={
              <span>
                <UserOutlined />
                作者信息
              </span>
            }
            key="1"
          >
            <Card bordered={false} className="tab-card">
              <Row gutter={[32, 16]}>
                <Col xs={24} md={8}>
                  <div className="field-preview">
                    <Avatar
                      src={config.AUTHOR_AVATAR || '/placeholder-avatar.png'}
                      size={150}
                      style={{
                        boxShadow: '0 4px 12px rgba(0, 0, 0, 0.1)',
                        border: '5px solid #fff',
                      }}
                    />
                    <div className="edit-overlay">
                      <EditOutlined />
                    </div>
                  </div>
                </Col>
                <Col xs={24} md={16}>
                  <Form.Item
                    name="AUTHOR_AVATAR"
                    label="作者头像"
                    tooltip="建议使用正方形图片，尺寸不小于 200x200"
                    extra="输入图片URL地址，建议使用CDN加速的图片链接"
                  >
                    <Input
                      prefix={<PictureOutlined />}
                      placeholder="请输入头像URL"
                      onChange={(e) =>
                        onChangeHandle('AUTHOR_AVATAR', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={12}>
                  <Form.Item
                    name="AUTHOR_NAME"
                    label="作者姓名"
                    tooltip="显示在个人资料和文章作者处"
                  >
                    <Input
                      prefix={<UserOutlined />}
                      placeholder="请输入作者姓名"
                      onChange={(e) =>
                        onChangeHandle('AUTHOR_NAME', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={12}>
                  <Form.Item
                    name="WEBSITE_AUTHOR"
                    label="网站作者"
                    tooltip="显示在网站元数据中，用于SEO"
                  >
                    <Input
                      prefix={<UserOutlined />}
                      placeholder="请输入网站作者"
                      onChange={(e) =>
                        onChangeHandle('WEBSITE_AUTHOR', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col span={24}>
                  <Form.Item
                    name="AUTHOR_INFO"
                    label="作者简介"
                    tooltip="简短的个人介绍，显示在个人资料页"
                  >
                    <Input.TextArea
                      placeholder="请输入作者简介"
                      autoSize={{ minRows: 2, maxRows: 4 }}
                      onChange={(e) =>
                        onChangeHandle('AUTHOR_INFO', e.target.value)
                      }
                      showCount
                      maxLength={200}
                    />
                  </Form.Item>
                </Col>

                <Col span={24}>
                  <Form.Item
                    name="AUTHOR_WELCOME"
                    label="欢迎信息"
                    tooltip="显示在首页或个人资料页的欢迎语"
                  >
                    <Input.TextArea
                      placeholder="请输入欢迎信息"
                      autoSize={{ minRows: 2, maxRows: 4 }}
                      onChange={(e) =>
                        onChangeHandle('AUTHOR_WELCOME', e.target.value)
                      }
                      showCount
                      maxLength={200}
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={12}>
                  <Form.Item
                    name="AUTHOR_GITHUB"
                    label="作者 GitHub"
                    tooltip="GitHub个人主页链接"
                  >
                    <Input
                      prefix={<GithubOutlined />}
                      placeholder="请输入GitHub链接"
                      onChange={(e) =>
                        onChangeHandle('AUTHOR_GITHUB', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={12}>
                  <Form.Item
                    name="AUTHOR_HOME"
                    label="作者主页"
                    tooltip="个人网站或博客链接"
                  >
                    <Input
                      prefix={<HomeOutlined />}
                      placeholder="请输入个人主页链接"
                      onChange={(e) =>
                        onChangeHandle('AUTHOR_HOME', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>
              </Row>
            </Card>
          </TabPane>

          <TabPane
            tab={
              <span>
                <GlobalOutlined />
                网站信息
              </span>
            }
            key="2"
          >
            <Card bordered={false} className="tab-card">
              <Row gutter={[32, 16]}>
                <Col xs={24} md={12}>
                  <div className="field-preview website-logo-preview">
                    {config.WEBSITE_LOGO ? (
                      <img
                        src={config.WEBSITE_LOGO || '/placeholder.svg'}
                        alt="网站Logo"
                        style={{ maxHeight: 80, maxWidth: '100%' }}
                      />
                    ) : (
                      <Empty
                        description="暂无Logo"
                        image={Empty.PRESENTED_IMAGE_SIMPLE}
                      />
                    )}
                    <div className="edit-overlay">
                      <EditOutlined />
                    </div>
                  </div>
                </Col>

                <Col xs={24} md={12}>
                  <div className="field-preview website-favicon-preview">
                    {config.WEBSITE_FAVICON ? (
                      <img
                        src={config.WEBSITE_FAVICON || '/placeholder.svg'}
                        alt="网站图标"
                        style={{ maxHeight: 60, maxWidth: '100%' }}
                      />
                    ) : (
                      <Empty
                        description="暂无图标"
                        image={Empty.PRESENTED_IMAGE_SIMPLE}
                      />
                    )}
                    <div className="edit-overlay">
                      <EditOutlined />
                    </div>
                  </div>
                </Col>

                <Col xs={24} md={12}>
                  <Form.Item
                    name="WEBSITE_LOGO"
                    label="网站 Logo"
                    tooltip="显示在网站顶部的Logo图片"
                    extra="建议使用透明背景的PNG图片"
                  >
                    <Input
                      prefix={<PictureOutlined />}
                      placeholder="请输入Logo URL"
                      onChange={(e) =>
                        onChangeHandle('WEBSITE_LOGO', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={12}>
                  <Form.Item
                    name="WEBSITE_FAVICON"
                    label="网站图标"
                    tooltip="显示在浏览器标签页的小图标"
                    extra="建议使用.ico格式或正方形PNG图片"
                  >
                    <Input
                      prefix={<PictureOutlined />}
                      placeholder="请输入Favicon URL"
                      onChange={(e) =>
                        onChangeHandle('WEBSITE_FAVICON', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={12}>
                  <Form.Item
                    name="WEBSITE_NAME"
                    label="网站名称"
                    required
                    tooltip="显示在浏览器标题栏和网站顶部"
                  >
                    <Input
                      prefix={<GlobalOutlined />}
                      placeholder="请输入网站名称"
                      onChange={(e) =>
                        onChangeHandle('WEBSITE_NAME', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={12}>
                  <Form.Item
                    name="WEBSITE_URL"
                    label="网站 URL"
                    tooltip="网站的完整访问地址，包含http(s)://"
                  >
                    <Input
                      prefix={<LinkOutlined />}
                      placeholder="请输入网站URL"
                      onChange={(e) =>
                        onChangeHandle('WEBSITE_URL', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col span={24}>
                  <Form.Item
                    name="WEBSITE_DESCRIPTION"
                    label="网站描述"
                    tooltip="用于SEO和社交媒体分享时的描述"
                  >
                    <Input.TextArea
                      placeholder="请输入网站描述"
                      autoSize={{ minRows: 2, maxRows: 4 }}
                      onChange={(e) =>
                        onChangeHandle('WEBSITE_DESCRIPTION', e.target.value)
                      }
                      showCount
                      maxLength={200}
                    />
                  </Form.Item>
                </Col>

                <Col span={24}>
                  <Form.Item
                    name="WEBSITE_KEYWORDS"
                    label="网站关键词"
                    tooltip="用于SEO，多个关键词用逗号分隔"
                  >
                    <Input
                      prefix={<TagsOutlined />}
                      placeholder="请输入网站关键词，用逗号分隔"
                      onChange={(e) =>
                        onChangeHandle('WEBSITE_KEYWORDS', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={8}>
                  <Form.Item
                    name="WEBSITE_CREATE_TIME"
                    label="网站创建时间"
                    tooltip="网站的创建或上线时间"
                  >
                    <Input
                      prefix={<FieldTimeOutlined />}
                      placeholder="请输入网站创建时间"
                      onChange={(e) =>
                        onChangeHandle('WEBSITE_CREATE_TIME', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={8}>
                  <Form.Item
                    name="WEBSITE_ICP"
                    label="ICP 备案号"
                    tooltip="中国大陆网站ICP备案号"
                  >
                    <Input
                      prefix={<SafetyCertificateOutlined />}
                      placeholder="请输入ICP备案号"
                      onChange={(e) =>
                        onChangeHandle('WEBSITE_ICP', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={8}>
                  <Form.Item
                    name="WEBSITE_MPS"
                    label="MPS 备案号"
                    tooltip="中国大陆网站公安备案号"
                  >
                    <Input
                      prefix={<SafetyCertificateOutlined />}
                      placeholder="请输入MPS备案号"
                      onChange={(e) =>
                        onChangeHandle('WEBSITE_MPS', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col span={24}>
                  <Form.Item
                    name="WEBSITE_COPYRIGHT"
                    label="网站版权"
                    tooltip="显示在网站底部的版权信息"
                  >
                    <Input
                      prefix={<CopyrightOutlined />}
                      placeholder="请输入网站版权信息"
                      onChange={(e) =>
                        onChangeHandle('WEBSITE_COPYRIGHT', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>
              </Row>
            </Card>
          </TabPane>

          <TabPane
            tab={
              <span>
                <HomeOutlined />
                首页设置
              </span>
            }
            key="3"
          >
            <Card bordered={false} className="tab-card">
              <Row gutter={[32, 16]}>
                <Col span={24}>
                  <div className="home-preview">
                    <div className="home-preview-content">
                      <Title level={3}>{config.HOME_TITLE || '主页标题'}</Title>
                      <Paragraph strong style={{ fontSize: 18 }}>
                        {config.HOME_SLOGAN || '主页标语将显示在这里'}
                      </Paragraph>
                      <Paragraph
                        italic
                        type="secondary"
                        style={{ fontSize: 16 }}
                      >
                        {config.HOME_SLOGAN_EN ||
                          'English slogan will be displayed here'}
                      </Paragraph>
                      {config.HOME_GITHUB && (
                        <Button
                          icon={<GithubOutlined />}
                          shape="round"
                          style={{ marginTop: 16 }}
                        >
                          GitHub
                        </Button>
                      )}
                    </div>
                  </div>
                </Col>

                <Col xs={24} md={12}>
                  <Form.Item
                    name="HOME_TITLE"
                    label="主页标题"
                    tooltip="显示在首页顶部的主标题"
                  >
                    <Input
                      placeholder="请输入主页标题"
                      onChange={(e) =>
                        onChangeHandle('HOME_TITLE', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} md={12}>
                  <Form.Item
                    name="HOME_GITHUB"
                    label="主页 GitHub"
                    tooltip="首页展示的GitHub仓库链接"
                  >
                    <Input
                      prefix={<GithubOutlined />}
                      placeholder="请输入GitHub链接"
                      onChange={(e) =>
                        onChangeHandle('HOME_GITHUB', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col span={24}>
                  <Form.Item
                    name="HOME_SLOGAN"
                    label="主页标语"
                    tooltip="显示在首页的中文标语"
                  >
                    <Input
                      placeholder="请输入主页标语"
                      onChange={(e) =>
                        onChangeHandle('HOME_SLOGAN', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>

                <Col span={24}>
                  <Form.Item
                    name="HOME_SLOGAN_EN"
                    label="主页标语（英文）"
                    tooltip="显示在首页的英文标语"
                  >
                    <Input
                      placeholder="请输入主页英文标语"
                      onChange={(e) =>
                        onChangeHandle('HOME_SLOGAN_EN', e.target.value)
                      }
                      allowClear
                    />
                  </Form.Item>
                </Col>
              </Row>
            </Card>
          </TabPane>
        </Tabs>

        <div style={{ display: 'none' }}>
          <Button type="primary" htmlType="submit">
            提交
          </Button>
        </div>
      </Form>

      {/* @ts-ignore */}
      <style jsx global>{`
        .loading-container {
          display: flex;
          justify-content: center;
          align-items: center;
          height: 400px;
        }

        .config-form {
          margin-top: 24px;
        }

        .config-tabs .ant-tabs-nav {
          margin-bottom: 16px;
        }

        .tab-card {
          box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03),
            0 1px 6px -1px rgba(0, 0, 0, 0.02), 0 2px 4px rgba(0, 0, 0, 0.02);
          border-radius: 8px;
        }

        .preview-section {
          margin-bottom: 32px;
        }

        .preview-card {
          height: 100%;
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
          border-radius: 8px;
          transition: all 0.3s ease;
          overflow: hidden;
        }

        .preview-card:hover {
          box-shadow: 0 6px 16px rgba(0, 0, 0, 0.08);
          transform: translateY(-2px);
        }

        .field-preview {
          display: flex;
          justify-content: center;
          align-items: center;
          padding: 16px;
          background-color: #fafafa;
          border-radius: 8px;
          position: relative;
          height: 100%;
          min-height: 120px;
          border: 1px dashed #d9d9d9;
          transition: all 0.3s;
        }

        .field-preview:hover {
          border-color: #1890ff;
          background-color: #f0f7ff;
        }

        .website-logo-preview,
        .website-favicon-preview {
          padding: 24px;
          display: flex;
          justify-content: center;
          align-items: center;
        }

        .edit-overlay {
          position: absolute;
          top: 8px;
          right: 8px;
          background-color: rgba(0, 0, 0, 0.5);
          color: white;
          width: 24px;
          height: 24px;
          border-radius: 50%;
          display: flex;
          justify-content: center;
          align-items: center;
          opacity: 0;
          transition: opacity 0.3s;
        }

        .field-preview:hover .edit-overlay {
          opacity: 1;
        }

        .home-preview {
          background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
          border-radius: 8px;
          padding: 32px;
          margin-bottom: 24px;
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
        }

        .home-preview-content {
          max-width: 600px;
          margin: 0 auto;
          text-align: center;
          padding: 32px 16px;
        }

        .ant-form-item-label > label {
          font-weight: 500;
        }

        .ant-tabs-tab.ant-tabs-tab-active .ant-tabs-tab-btn {
          font-weight: 500;
        }

        .ant-form-item-extra {
          font-size: 12px;
          color: #8c8c8c;
        }

        .ant-divider-inner-text {
          font-weight: 500;
        }
      `}</style>
    </PageContainer>
  );
};

export default ConfigPage;
