import AISummaryController from '@/services/aisummary/AISummaryController';
import {
  BookOutlined,
  CheckCircleOutlined,
  FileTextOutlined,
  LoadingOutlined,
  RobotOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import {
  Alert,
  Button,
  Card,
  Divider,
  Empty,
  Modal,
  Select,
  Space,
  Spin,
  Tag,
  Typography,
} from 'antd';
import React, { useEffect, useState } from 'react';

const { Option } = Select;
const { Title, Paragraph, Text } = Typography;

interface AiSummaryModalProps {
  visible: boolean;
  onClose: () => void;
  contentType: 'ARTICLE' | 'MOMENT' | 'PAGE';
  contentId: string;
}

const modelOptions = [
  {
    value: 'deepseek-chat',
    label: 'DeepSeek-V3',
    icon: <ThunderboltOutlined />,
  },
  { value: 'deepseek-reasoner', label: 'DeepSeek-R1', icon: <RobotOutlined /> },
];

const getContentIcon = (type: string) => {
  switch (type) {
    case 'ARTICLE':
      return <FileTextOutlined />;
    case 'MOMENT':
      return <BookOutlined />;
    default:
      return <FileTextOutlined />;
  }
};

const AiSummaryModal: React.FC<AiSummaryModalProps> = ({
  visible,
  onClose,
  contentType,
  contentId,
}) => {
  const [model, setModel] = useState<string>('deepseek-chat');
  const [summary, setSummary] = useState<string>('');
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [isFinished, setIsFinished] = useState<boolean>(false);

  const startStreaming = (taskId: string) => {
    const eventSource = AISummaryController.streamSummary(taskId);

    eventSource.onmessage = (event) => {
      const result = JSON.parse(event.data);
      console.log(new Date());
      console.log(result);
      setSummary(result.content);

      if (result.status === 'COMPLETED') {
        eventSource.close();
        setLoading(false);
        setIsFinished(true);
      }
    };

    eventSource.onerror = () => {
      eventSource.close();
      setError('出现错误了捏，请重试');
      setLoading(false);
    };
  };

  const handleCreateSummaryTask = async () => {
    setLoading(true);
    setError(null);
    setSummary('');
    setIsFinished(false);

    try {
      const response = await AISummaryController.createSummaryTask({
        type: contentType,
        targetId: contentId,
        model: model,
      });

      const taskId = response.data.taskId;
      startStreaming(taskId);
    } catch (err) {
      setError('创建任务失败，请重试');
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!visible) {
      setSummary('');
      setError(null);
      setIsFinished(false);
    }
  }, [visible]);

  const renderSummaryContent = () => {
    if (error) {
      return (
        <Alert
          message="生成失败"
          description={
            <Space>
              {error}
              <Button type="link" onClick={handleCreateSummaryTask}>
                重试
              </Button>
            </Space>
          }
          type="error"
          showIcon
        />
      );
    }

    if (!summary && !loading) {
      return (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="选择模型并点击生成按钮开始 AI 总结"
        />
      );
    }

    return (
      <Card
        className="summary-card"
        bordered={false}
        style={{
          backgroundColor: '#f9f9f9',
          borderRadius: '8px',
          marginTop: 16,
        }}
      >
        {isFinished ? (
          <div className="summary-status">
            <CheckCircleOutlined style={{ color: '#52c41a' }} />
            <Text type="success">总结已完成</Text>
          </div>
        ) : loading && summary ? (
          <div className="summary-status">
            <LoadingOutlined style={{ color: '#1890ff' }} />
            <Text type="secondary">正在生成中...</Text>
          </div>
        ) : null}

        <Paragraph
          style={{
            whiteSpace: 'pre-wrap',
            fontSize: '14px',
            lineHeight: '1.8',
            margin: 0,
          }}
        >
          {summary}
          {loading && (
            <Text type="secondary">
              <span className="typing-indicator">
                <span>.</span>
                <span>.</span>
                <span>.</span>
              </span>
            </Text>
          )}
        </Paragraph>
      </Card>
    );
  };

  return (
    <Modal
      title={
        <Space>
          <RobotOutlined />
          <span>AI 总结生成</span>
          <Tag color="blue">
            {getContentIcon(contentType)} {contentType}
          </Tag>
        </Space>
      }
      open={visible}
      onCancel={onClose}
      width={600}
      footer={[
        <Button key="close" onClick={onClose}>
          关闭
        </Button>,
        <Button
          key="generate"
          type="primary"
          icon={isFinished ? <CheckCircleOutlined /> : <ThunderboltOutlined />}
          loading={loading && !summary}
          onClick={isFinished ? onClose : handleCreateSummaryTask}
          disabled={loading && !isFinished}
        >
          {isFinished ? '完成' : '生成总结'}
        </Button>,
      ]}
      bodyStyle={{ padding: '16px 24px' }}
    >
      <div className="model-selection">
        <Title level={5}>选择 AI 模型</Title>
        <Select
          style={{ width: '100%' }}
          value={model}
          onChange={(value) => setModel(value)}
          optionLabelProp="label"
          disabled={loading}
        >
          {modelOptions.map((option) => (
            <Option
              key={option.value}
              value={option.value}
              label={option.label}
            >
              <Space>
                {option.icon}
                <Text strong>{option.label}</Text>
              </Space>
            </Option>
          ))}
        </Select>
      </div>

      <Divider />

      <div className="summary-section">
        <Title level={5} style={{ display: 'flex', alignItems: 'center' }}>
          {summary || loading ? '总结内容' : '生成结果'}
          {loading && !summary && (
            <Spin size="small" style={{ marginLeft: 8 }} />
          )}
        </Title>

        {!summary && loading ? (
          <div className="summary-loading">
            <Spin tip="AI 正在准备生成..." />
          </div>
        ) : (
          renderSummaryContent()
        )}
      </div>

      <style jsx>{`
        .summary-loading {
          display: flex;
          justify-content: center;
          align-items: center;
          padding: 40px 0;
        }

        .summary-status {
          display: flex;
          align-items: center;
          gap: 8px;
          margin-bottom: 12px;
        }

        .model-selection,
        .summary-section {
          margin-bottom: 16px;
        }

        .typing-indicator {
          display: inline-block;
        }

        .typing-indicator span {
          animation: blink 1s infinite;
          animation-fill-mode: both;
        }

        .typing-indicator span:nth-child(2) {
          animation-delay: 0.2s;
        }

        .typing-indicator span:nth-child(3) {
          animation-delay: 0.4s;
        }

        @keyframes blink {
          0% {
            opacity: 0.2;
          }
          20% {
            opacity: 1;
          }
          100% {
            opacity: 0.2;
          }
        }
      `}</style>
    </Modal>
  );
};

export default AiSummaryModal;
