'use client';

import {
  BgColorsOutlined,
  CheckCircleOutlined,
  EditOutlined,
  PictureOutlined,
  RocketOutlined,
  SettingOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { history } from '@umijs/max';
import {
  Avatar,
  Button,
  Card,
  Col,
  Divider,
  Form,
  Input,
  message,
  Radio,
  Result,
  Row,
  Select,
  Steps,
  Typography,
  Upload,
} from 'antd';
import type React from 'react';
import { useEffect, useState } from 'react';
import styles from './index.less';

const { Title, Paragraph, Text } = Typography;
const { Step } = Steps;
const { Option } = Select;

const BlogInitializePage: React.FC = () => {
  const [current, setCurrent] = useState(0);
  const [form] = Form.useForm();
  const [blogConfig, setBlogConfig] = useState({
    blogName: '',
    blogDescription: '',
    username: '',
    email: '',
    avatar: null,
    theme: 'light',
  });
  const [selectedTheme, setSelectedTheme] = useState('light');
  const [animateStep, setAnimateStep] = useState(false);

  // 添加动画效果
  useEffect(() => {
    setAnimateStep(false);
    const timer = setTimeout(() => {
      setAnimateStep(true);
    }, 100);
    return () => clearTimeout(timer);
  }, [current]);

  const next = () => {
    form.validateFields().then((values) => {
      setCurrent(current + 1);
      setBlogConfig({ ...blogConfig, ...values });
    });
  };

  const prev = () => {
    setCurrent(current - 1);
  };

  const handleFinish = () => {
    message.success('博客初始化成功！即将跳转到管理页面...');
    // 在实际应用中，这里会提交所有配置到后端
    setTimeout(() => {
      history.push('/dashboard');
    }, 1500);
  };

  const steps = [
    {
      title: '欢迎',
      icon: <RocketOutlined />,
      content: (
        <Card
          className={`${styles.welcomeCard} ${
            animateStep ? styles.fadeIn : ''
          }`}
          bordered={false}
        >
          <div className={styles.welcomeContent}>
            <div className={styles.welcomeHeader}>
              <div className={styles.welcomeLogo}>
                <div className={styles.logoCircle}>
                  <EditOutlined className={styles.logoIcon} />
                </div>
              </div>
              <Title level={2} className={styles.welcomeTitle}>
                欢迎使用 Grtblog
              </Title>
              <Paragraph className={styles.welcomeSubtitle}>
                只需几个简单步骤，开启您的创作之旅
              </Paragraph>
            </div>
            {/*<div className={styles.welcomeImageContainer}>*/}
            {/*  <div className={styles.welcomeImage}>*/}
            {/*    <img*/}
            {/*      src="/placeholder.svg?height=240&width=480"*/}
            {/*      alt="欢迎使用博客系统"*/}
            {/*    />*/}
            {/*  </div>*/}
            {/*  <div className={styles.imageOverlay}></div>*/}
            {/*</div>*/}
            <div className={styles.featuresList}>
              <Row gutter={[24, 24]}>
                <Col xs={24} sm={12}>
                  <div className={styles.featureItem}>
                    <div className={styles.featureIcon}>
                      <SettingOutlined />
                    </div>
                    <div className={styles.featureText}>
                      <Text strong>博客设置</Text>
                      <Text type="secondary">自定义您的博客信息</Text>
                    </div>
                  </div>
                </Col>
                <Col xs={24} sm={12}>
                  <div className={styles.featureItem}>
                    <div className={styles.featureIcon}>
                      <UserOutlined />
                    </div>
                    <div className={styles.featureText}>
                      <Text strong>个人资料</Text>
                      <Text type="secondary">创建您的管理员账户</Text>
                    </div>
                  </div>
                </Col>
                <Col xs={24} sm={12}>
                  <div className={styles.featureItem}>
                    <div className={styles.featureIcon}>
                      <BgColorsOutlined />
                    </div>
                    <div className={styles.featureText}>
                      <Text strong>主题选择</Text>
                      <Text type="secondary">选择您喜欢的视觉风格</Text>
                    </div>
                  </div>
                </Col>
                <Col xs={24} sm={12}>
                  <div className={styles.featureItem}>
                    <div className={styles.featureIcon}>
                      <RocketOutlined />
                    </div>
                    <div className={styles.featureText}>
                      <Text strong>开始创作</Text>
                      <Text type="secondary">立即开始您的创作之旅</Text>
                    </div>
                  </div>
                </Col>
              </Row>
            </div>
          </div>
        </Card>
      ),
    },
    {
      title: '博客设置',
      icon: <SettingOutlined />,
      content: (
        <Card
          title={
            <div className={styles.stepTitle}>
              <SettingOutlined className={styles.stepIcon} />
              <span>博客基本设置</span>
            </div>
          }
          className={`${styles.formCard} ${animateStep ? styles.fadeIn : ''}`}
          bordered={false}
        >
          <Form
            form={form}
            layout="vertical"
            initialValues={{
              blogName: blogConfig.blogName,
              blogDescription: blogConfig.blogDescription,
            }}
          >
            <Form.Item
              name="blogName"
              label="博客名称"
              rules={[{ required: true, message: '请输入博客名称' }]}
            >
              <Input
                placeholder="输入您的博客名称"
                prefix={<EditOutlined />}
                className={styles.formInput}
              />
            </Form.Item>
            <Form.Item
              name="blogDescription"
              label="博客描述"
              rules={[{ required: true, message: '请输入博客描述' }]}
            >
              <Input.TextArea
                rows={4}
                placeholder="简单描述一下您的博客"
                className={styles.formTextarea}
              />
            </Form.Item>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="language"
                  label="默认语言"
                  initialValue="zh-CN"
                >
                  <Select className={styles.formSelect}>
                    <Option value="zh-CN">简体中文</Option>
                    <Option value="en-US">English</Option>
                    <Option value="ja-JP">日本語</Option>
                  </Select>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  name="visibility"
                  label="博客可见性"
                  initialValue="public"
                >
                  <Select className={styles.formSelect}>
                    <Option value="public">公开</Option>
                    <Option value="private">私密</Option>
                    <Option value="password">密码保护</Option>
                  </Select>
                </Form.Item>
              </Col>
            </Row>
          </Form>
        </Card>
      ),
    },
    {
      title: '个人资料',
      icon: <UserOutlined />,
      content: (
        <Card
          title={
            <div className={styles.stepTitle}>
              <UserOutlined className={styles.stepIcon} />
              <span>管理员账户设置</span>
            </div>
          }
          className={`${styles.formCard} ${animateStep ? styles.fadeIn : ''}`}
          bordered={false}
        >
          <div className={styles.avatarSection}>
            <Avatar
              size={80}
              icon={<UserOutlined />}
              className={styles.avatarPreview}
            />
            <div className={styles.avatarUpload}>
              <Upload
                name="avatar"
                listType="picture-card"
                className="avatar-uploader"
                showUploadList={false}
                beforeUpload={() => false}
              >
                <div className={styles.uploadButton}>
                  <PictureOutlined className={styles.uploadIcon} />
                  <div className={styles.uploadText}>上传头像</div>
                </div>
              </Upload>
            </div>
          </div>
          <Form
            form={form}
            layout="vertical"
            initialValues={{
              username: blogConfig.username,
              email: blogConfig.email,
            }}
          >
            <Form.Item
              name="username"
              label="用户名"
              rules={[{ required: true, message: '请输入用户名' }]}
            >
              <Input
                prefix={<UserOutlined />}
                placeholder="设置管理员用户名"
                className={styles.formInput}
              />
            </Form.Item>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item
                  name="email"
                  label="电子邮箱"
                  rules={[
                    { required: true, message: '请输入电子邮箱' },
                    { type: 'email', message: '请输入有效的电子邮箱' },
                  ]}
                >
                  <Input
                    placeholder="您的电子邮箱"
                    className={styles.formInput}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="nickname" label="昵称">
                  <Input
                    placeholder="您在博客中显示的名称"
                    className={styles.formInput}
                  />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item
              name="password"
              label="密码"
              rules={[{ required: true, message: '请设置密码' }]}
            >
              <Input.Password
                placeholder="设置安全密码"
                className={styles.formInput}
              />
            </Form.Item>
            <Form.Item name="bio" label="个人简介">
              <Input.TextArea
                rows={3}
                placeholder="简单介绍一下自己"
                className={styles.formTextarea}
              />
            </Form.Item>
          </Form>
        </Card>
      ),
    },
    {
      title: '主题选择',
      icon: <BgColorsOutlined />,
      content: (
        <Card
          title={
            <div className={styles.stepTitle}>
              <BgColorsOutlined className={styles.stepIcon} />
              <span>选择博客主题</span>
            </div>
          }
          className={`${styles.themeCard} ${animateStep ? styles.fadeIn : ''}`}
          bordered={false}
        >
          <Paragraph className={styles.themeDescription}>
            选择一个适合您风格的主题，您随时可以在管理后台更改主题设置。
          </Paragraph>
          <Form form={form}>
            <Form.Item name="theme" initialValue={selectedTheme}>
              <Radio.Group
                onChange={(e) => setSelectedTheme(e.target.value)}
                value={selectedTheme}
                className={styles.themeRadioGroup}
              >
                <Row gutter={[24, 24]}>
                  <Col xs={24} md={8}>
                    <Radio value="light" className={styles.themeRadio}>
                      <Card
                        className={`${styles.themeOption} ${
                          selectedTheme === 'light' ? styles.selectedTheme : ''
                        }`}
                        hoverable
                      >
                        <div
                          className={styles.themePreview}
                          style={{ background: '#ffffff', color: '#000000' }}
                        >
                          <div
                            className={styles.themeHeader}
                            style={{ background: '#f0f2f5' }}
                          >
                            明亮主题
                          </div>
                          <div className={styles.themeContent}>
                            <div
                              className={styles.themeSidebar}
                              style={{
                                background: '#ffffff',
                                borderRight: '1px solid #f0f0f0',
                              }}
                            ></div>
                            <div className={styles.themeMain}></div>
                          </div>
                        </div>
                        <div className={styles.themeInfo}>
                          <div className={styles.themeName}>明亮主题</div>
                          <div className={styles.themeDesc}>
                            清新简约的设计风格
                          </div>
                        </div>
                        {selectedTheme === 'light' && (
                          <div className={styles.themeSelected}>
                            <CheckCircleOutlined />
                          </div>
                        )}
                      </Card>
                    </Radio>
                  </Col>
                  <Col xs={24} md={8}>
                    <Radio value="dark" className={styles.themeRadio}>
                      <Card
                        className={`${styles.themeOption} ${
                          selectedTheme === 'dark' ? styles.selectedTheme : ''
                        }`}
                        hoverable
                      >
                        <div
                          className={styles.themePreview}
                          style={{ background: '#141414', color: '#ffffff' }}
                        >
                          <div
                            className={styles.themeHeader}
                            style={{ background: '#1f1f1f' }}
                          >
                            暗黑主题
                          </div>
                          <div className={styles.themeContent}>
                            <div
                              className={styles.themeSidebar}
                              style={{
                                background: '#141414',
                                borderRight: '1px solid #303030',
                              }}
                            ></div>
                            <div className={styles.themeMain}></div>
                          </div>
                        </div>
                        <div className={styles.themeInfo}>
                          <div className={styles.themeName}>暗黑主题</div>
                          <div className={styles.themeDesc}>
                            护眼且富有科技感
                          </div>
                        </div>
                        {selectedTheme === 'dark' && (
                          <div className={styles.themeSelected}>
                            <CheckCircleOutlined />
                          </div>
                        )}
                      </Card>
                    </Radio>
                  </Col>
                  <Col xs={24} md={8}>
                    <Radio value="colorful" className={styles.themeRadio}>
                      <Card
                        className={`${styles.themeOption} ${
                          selectedTheme === 'colorful'
                            ? styles.selectedTheme
                            : ''
                        }`}
                        hoverable
                      >
                        <div
                          className={styles.themePreview}
                          style={{ background: '#ffffff', color: '#000000' }}
                        >
                          <div
                            className={styles.themeHeader}
                            style={{ background: '#1890ff', color: '#ffffff' }}
                          >
                            多彩主题
                          </div>
                          <div className={styles.themeContent}>
                            <div
                              className={styles.themeSidebar}
                              style={{
                                background: '#ffffff',
                                borderRight: '1px solid #f0f0f0',
                              }}
                            ></div>
                            <div className={styles.themeMain}></div>
                          </div>
                        </div>
                        <div className={styles.themeInfo}>
                          <div className={styles.themeName}>多彩主题</div>
                          <div className={styles.themeDesc}>
                            丰富多彩的视觉体验
                          </div>
                        </div>
                        {selectedTheme === 'colorful' && (
                          <div className={styles.themeSelected}>
                            <CheckCircleOutlined />
                          </div>
                        )}
                      </Card>
                    </Radio>
                  </Col>
                </Row>
              </Radio.Group>
            </Form.Item>
          </Form>
        </Card>
      ),
    },
    {
      title: '完成',
      icon: <CheckCircleOutlined />,
      content: (
        <Card
          className={`${styles.completeCard} ${
            animateStep ? styles.fadeIn : ''
          }`}
          bordered={false}
        >
          <Result
            status="success"
            icon={
              <div className={styles.successIcon}>
                <CheckCircleOutlined />
              </div>
            }
            title={<div className={styles.successTitle}>博客初始化成功！</div>}
            subTitle={
              <div className={styles.successSubtitle}>
                您的博客已经准备就绪，现在可以开始您的创作之旅了。
              </div>
            }
            extra={[
              <Button
                type="primary"
                key="console"
                onClick={handleFinish}
                size="large"
                className={styles.successButton}
              >
                进入管理后台
              </Button>,
            ]}
          />
          <div className={styles.completeTips}>
            <Divider>
              <span className={styles.tipsTitle}>快速开始</span>
            </Divider>
            <Row gutter={[16, 16]} className={styles.tipsList}>
              <Col xs={24} sm={8}>
                <Card className={styles.tipCard}>
                  <EditOutlined className={styles.tipIcon} />
                  <div className={styles.tipTitle}>发布第一篇文章</div>
                  <div className={styles.tipDesc}>分享您的想法和创意</div>
                </Card>
              </Col>
              <Col xs={24} sm={8}>
                <Card className={styles.tipCard}>
                  <SettingOutlined className={styles.tipIcon} />
                  <div className={styles.tipTitle}>自定义博客</div>
                  <div className={styles.tipDesc}>个性化您的博客设置</div>
                </Card>
              </Col>
              <Col xs={24} sm={8}>
                <Card className={styles.tipCard}>
                  <UserOutlined className={styles.tipIcon} />
                  <div className={styles.tipTitle}>邀请用户</div>
                  <div className={styles.tipDesc}>邀请朋友加入您的博客</div>
                </Card>
              </Col>
            </Row>
          </div>
        </Card>
      ),
    },
  ];

  return (
    <PageContainer title={false} ghost>
      <div className={styles.container}>
        <div className={styles.initContent}>
          <Card className={styles.stepsCard} bordered={false}>
            <Steps current={current} className={styles.steps}>
              {steps.map((item) => (
                <Step key={item.title} title={item.title} icon={item.icon} />
              ))}
            </Steps>
          </Card>
          <div className={styles.stepsContent}>{steps[current].content}</div>
          <div className={styles.stepsAction}>
            {current > 0 && (
              <Button className={styles.prevButton} onClick={prev}>
                上一步
              </Button>
            )}
            {current < steps.length - 1 && (
              <Button
                type="primary"
                onClick={next}
                className={styles.nextButton}
              >
                下一步
              </Button>
            )}
            {current === steps.length - 1 && (
              <Button
                type="primary"
                onClick={handleFinish}
                className={styles.finishButton}
              >
                完成
              </Button>
            )}
          </div>
        </div>
      </div>
    </PageContainer>
  );
};

export default BlogInitializePage;
