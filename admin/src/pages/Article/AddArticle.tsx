import React, { useEffect, useRef, useState } from 'react';
import { Editor } from '@toast-ui/react-editor';
import '@toast-ui/editor/dist/toastui-editor.css';
import { Form, Input, message, Select, Upload } from 'antd';
import { AddArticleApiParams } from '@/services/article/typings';
import { ProForm } from '@ant-design/pro-components';
import { PlusOutlined } from '@ant-design/icons';
import ArticleController from '@/services/article/ArticleController';

const AddArticle = () => {
  const editorRef = useRef<Editor>(null);

  useEffect(() => {
    console.log('123');
  }, []);

  const [form, setForm] = useState<AddArticleApiParams>({
    title: '',
    content: '',
    cover: '',
    categoryId: 0,
    tags: '',
    status: '',
    createdAt: '',
    updatedAt: '',
  });
  const onValueChange = (key: string, value: any) => {
    setForm({
      ...form,
      [key]: value,
    });
  };

  const submitHandle = () => {
    setForm({
      ...form,
      content: editorRef.current?.getInstance().getMarkdown(),
    });
    console.log(form);
    // 检查一下是否有空的字段，除了封面
    if (!form.title || !form.content || !form.categoryId || !form.tags || !form.status) {
      message.error('请填写完整的文章信息');
      return;
    }
    ArticleController.addArticle(form).then((res) => {
      console.log(res);
    });
  };
  return (
    <div>
      {/* 标题 */}
      <ProForm
        title="添加文章"
        size="large"
        style={{ marginBottom: '20px' }}
        onFinish={submitHandle}
      >
        <Form.Item label="标题" name="title"
                   rules={[{ required: true, message: '请输入标题' }]}>
          <Input
            placeholder="标题"
            onChange={(e) => onValueChange('title', e.target.value)}
          />
        </Form.Item>
        <Form.Item label={'内容'} name="content"
                   rules={[{ required: true, message: '文章内容不能为空' }]}>
          <Editor
            previewStyle="vertical"
            height="600px"
            useCommandShortcut={true}
            initialEditType="markdown"
            initialValue=""
            ref={editorRef}
          />
        </Form.Item>
        <Form.Item>
          <div style={{ marginTop: '20px', marginRight: '20px' }}> 当然，你也可以选择从本地 markdown 文件导入</div>
          <Upload
            action="/api/upload"
            onChange={(e) => {
              if (e.file.status === 'done') {
                const url = e.file.response.data;
                editorRef.current?.getInstance().setMarkdown(url);
              }
            }}
          >
            <a> 点击上传 </a>
          </Upload>
        </Form.Item>
        <Form.Item label="上传封面">
          <Upload
            listType="picture-card"
            maxCount={1}
            action="/api/upload"
            onChange={(e) => {
              if (e.file.status === 'done') {
                const url = e.file.response.data;
                onValueChange('cover', url);
              }
            }}
          >
            <div>
              <PlusOutlined />
              <div style={{ marginTop: '8px' }}> 封面可选</div>
            </div>
          </Upload>
        </Form.Item>
        <Form.Item label="选择分类" name="category"
                   rules={[{ required: true, message: '请选择分类' }]}>
          <Select
            showSearch
            style={{ width: 200 }}
            placeholder="选择文章的分类"
            optionFilterProp="children"
            onChange={(value) => onValueChange('categoryId', value)}
          >
            <Select.Option value="0"> JavaScript </Select.Option>
            <Select.Option value="1"> TypeScript </Select.Option>
            <Select.Option value="2"> React </Select.Option>
            <Select.Option value="3"> Vue </Select.Option>
          </Select>
        </Form.Item>
        <Form.Item
          label="标签"
          name="tags"
          rules={[{ required: true, message: '请输入标签' }]}
        >
          <Input placeholder="请输入标签，使用英文逗号分隔"
                 onChange={(e) => onValueChange('tags', e.target.value)}
          />
        </Form.Item>
        <Form.Item label="状态" name="status"
                   rules={[{ required: true, message: '请选择文章状态' }]}>
          <Select style={{ width: 200 }} placeholder="选择文章的状态"
                  onChange={(value) => onValueChange('status', value)}
          >
            <Select.Option value="DRAFT"> 草稿 </Select.Option>
            <Select.Option value="PUBLISHED"> 发布 </Select.Option>
          </Select>
        </Form.Item>
        <Form.Item label="创建时间" name="createdAt">
          <Input
            placeholder="创建时间"
            onChange={(e) => onValueChange('createdAt', e.target.value)}
          />
        </Form.Item>
        <Form.Item label="更新时间" name="updatedAt">
          <Input
            placeholder="更新时间"
            onChange={(e) => onValueChange('updatedAt', e.target.value)}
          />
        </Form.Item>
      </ProForm>
    </div>
  );
};

export default AddArticle;
