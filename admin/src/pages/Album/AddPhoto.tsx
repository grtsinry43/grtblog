'use client';

import AlbumController from '@/services/album/AlbumController';
import {
  CameraOutlined,
  ClockCircleOutlined,
  EnvironmentOutlined,
  FileTextOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import {
  Button,
  Card,
  Col,
  DatePicker,
  Form,
  Input,
  message,
  Row,
  Space,
  Typography,
  Upload,
} from 'antd';
import EXIF from 'exif-js';
import moment, { type Moment } from 'moment';
import { useRef, useState } from 'react';

const { Title, Text } = Typography;

const AddPhoto = () => {
  const [photoInfo, setPhotoInfo] = useState({
    url: '',
    device: '',
    location: '',
    description: '',
    time: null as Moment | null,
    shade: '',
  });
  const [loading, setLoading] = useState(false);
  const [previewImage, setPreviewImage] = useState('');

  const formRef = useRef<any>(null);

  const onValueChange = (key: string, value: any) => {
    setPhotoInfo((prevPhotoInfo) => ({
      ...prevPhotoInfo,
      [key]: value,
    }));
  };

  const getMainColor = (img: HTMLImageElement) => {
    const canvas = document.createElement('canvas');
    const context = canvas.getContext('2d');
    if (!context) return '';

    canvas.width = img.width;
    canvas.height = img.height;
    context.drawImage(img, 0, 0, img.width, img.height);

    const imageData = context.getImageData(0, 0, img.width, img.height);
    const data = imageData.data;
    const length = data.length;
    const color = { r: 0, g: 0, b: 0 };
    let count = 0;

    for (let i = 0; i < length; i += 4) {
      color.r += data[i];
      color.g += data[i + 1];
      color.b += data[i + 2];
      count++;
    }

    color.r = Math.floor(color.r / count);
    color.g = Math.floor(color.g / count);
    color.b = Math.floor(color.b / count);

    // Ensure proper hex format with padding
    const r = color.r.toString(16).padStart(2, '0');
    const g = color.g.toString(16).padStart(2, '0');
    const b = color.b.toString(16).padStart(2, '0');

    return `#${r}${g}${b}`;
  };

  const handleFileChange = (e: any) => {
    if (e.file.status === 'uploading') {
      setLoading(true);
      return;
    }

    if (e.file.status === 'done') {
      setLoading(false);
      const url = e.file.response.data;
      setPreviewImage(location.protocol + '//' + location.host + url);

      setPhotoInfo((prevPhotoInfo) => ({
        ...prevPhotoInfo,
        url,
      }));

      const img = new Image();
      img.crossOrigin = 'anonymous';
      img.src = location.protocol + '//' + location.host + url;

      img.onload = () => {
        // @ts-ignore
        EXIF.getData(img, function (this: any) {
          const exifData = EXIF.getAllTags(this);
          const dateTimeOriginal = exifData.DateTimeOriginal
            ? moment(exifData.DateTimeOriginal, 'YYYY:MM:DD HH:mm:ss')
            : null;

          const deviceInfo =
            exifData.Make && exifData.Model
              ? `${exifData.Make} ${exifData.Model}`
              : '';

          const locationInfo =
            exifData.GPSLatitude && exifData.GPSLongitude
              ? `${exifData.GPSLatitude}, ${exifData.GPSLongitude}`
              : '';

          formRef.current?.setFieldsValue({
            device: deviceInfo,
            time: dateTimeOriginal,
            location: locationInfo,
          });

          const mainColor = getMainColor(img);

          setPhotoInfo((prevPhotoInfo) => ({
            ...prevPhotoInfo,
            device: deviceInfo,
            time: dateTimeOriginal,
            location: locationInfo,
            shade: mainColor,
          }));
        });
      };
    }

    if (e.file.status === 'error') {
      setLoading(false);
      message.error('上传失败，请重试');
    }
  };

  const submitHandle = () => {
    formRef.current?.validateFields().then(() => {
      setLoading(true);

      AlbumController.uploadPhoto({
        ...photoInfo,
        time: photoInfo.time?.format('YYYY-MM-DDTHH:mm:ss') || '',
      })
        .then((response) => {
          if (response.data) {
            message.success('上传成功');
            formRef.current?.resetFields();
            setPhotoInfo({
              url: '',
              device: '',
              location: '',
              description: '',
              time: null,
              shade: '',
            });
            setPreviewImage('');
          } else {
            message.error(response.msg || '上传失败');
          }
        })
        .catch((err) => {
          message.error('上传失败: ' + (err.message || '未知错误'));
        })
        .finally(() => {
          setLoading(false);
        });
    });
  };

  const uploadButton = (
    <div>
      <PlusOutlined />
      <div style={{ marginTop: 8 }}>上传照片</div>
    </div>
  );

  return (
    <PageContainer title="添加照片" subTitle="上传照片并填写相关信息">
      <Card bordered={false} className="photo-upload-card">
        <Row gutter={[24, 0]}>
          <Col xs={24} md={12}>
            <div className="upload-preview-container">
              <Form.Item label="上传照片" required>
                <Upload
                  name="file"
                  listType="picture-card"
                  maxCount={1}
                  action="/api/v1/upload"
                  onChange={handleFileChange}
                  showUploadList={false}
                >
                  {previewImage ? (
                    <img
                      src={previewImage || '/placeholder.svg'}
                      alt="预览图"
                      style={{
                        width: '100%',
                        height: '100%',
                        objectFit: 'cover',
                      }}
                    />
                  ) : (
                    uploadButton
                  )}
                </Upload>
              </Form.Item>

              {photoInfo.shade && (
                <div className="color-preview">
                  <Text>主色调: {photoInfo.shade}</Text>
                  <div
                    className="color-box"
                    style={{
                      backgroundColor: photoInfo.shade,
                      width: '24px',
                      height: '24px',
                      marginLeft: '8px',
                      borderRadius: '4px',
                      border: '1px solid #d9d9d9',
                    }}
                  />
                </div>
              )}
            </div>
          </Col>

          <Col xs={24} md={12}>
            <Form
              ref={formRef}
              layout="vertical"
              onFinish={submitHandle}
              initialValues={photoInfo}
            >
              <Form.Item
                label={
                  <Space>
                    <CameraOutlined /> 设备名
                  </Space>
                }
                name="device"
                rules={[{ required: true, message: '请输入设备名' }]}
              >
                <Input
                  placeholder="相机/手机型号"
                  value={photoInfo.device}
                  onChange={(e) => onValueChange('device', e.target.value)}
                />
              </Form.Item>

              <Form.Item
                label={
                  <Space>
                    <EnvironmentOutlined /> 拍摄地点
                  </Space>
                }
                name="location"
                rules={[{ required: true, message: '请输入拍摄地点' }]}
              >
                <Input
                  placeholder="拍摄地点"
                  value={photoInfo.location}
                  onChange={(e) => onValueChange('location', e.target.value)}
                />
              </Form.Item>

              <Form.Item
                label={
                  <Space>
                    <FileTextOutlined /> 图片描述
                  </Space>
                }
                name="description"
                rules={[{ required: true, message: '请输入图片描述' }]}
              >
                <Input.TextArea
                  rows={4}
                  placeholder="描述这张照片..."
                  value={photoInfo.description}
                  onChange={(e) => onValueChange('description', e.target.value)}
                />
              </Form.Item>

              <Form.Item
                label={
                  <Space>
                    <ClockCircleOutlined /> 拍摄时间
                  </Space>
                }
                name="time"
                rules={[{ required: true, message: '请选择时间' }]}
              >
                <DatePicker
                  showTime
                  style={{ width: '100%' }}
                  placeholder="选择拍摄时间"
                  value={photoInfo.time}
                  onChange={(date) => onValueChange('time', date)}
                />
              </Form.Item>

              <Form.Item>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={loading}
                  block
                  size="large"
                >
                  提交
                </Button>
              </Form.Item>
            </Form>
          </Col>
        </Row>
      </Card>

      <style jsx global>{`
        .photo-upload-card {
          box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
        }

        .upload-preview-container {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 100%;
          padding: 16px;
        }

        .color-preview {
          display: flex;
          align-items: center;
          margin-top: 16px;
          padding: 8px 16px;
          background-color: #f5f5f5;
          border-radius: 4px;
        }

        .ant-upload.ant-upload-select-picture-card {
          width: 240px;
          height: 240px;
          margin: 0 auto;
        }

        @media (max-width: 768px) {
          .ant-upload.ant-upload-select-picture-card {
            width: 180px;
            height: 180px;
          }
        }
      `}</style>
    </PageContainer>
  );
};

export default AddPhoto;
