import { getAllConfig, updateConfig } from '@/services/config/ConfigController';
import { PageContainer } from '@ant-design/pro-components';
import { Button, Form, Input, message } from 'antd';
import React, { useEffect, useState } from 'react';
import {refreshFrontendCache} from "@/services/refersh";
import useRouteLeaveConfirm from "@/hooks/use-route-leave-confirm";

const ConfigPage = () => {
  const [config, setConfig] = useState({});
  const formRef = React.useRef<any>(null);
  useRouteLeaveConfirm();

  useEffect(() => {
    getAllConfig().then((res) => {
      if (res && res.data) {
        formRef.current.setFieldsValue(res.data);
      }
    });
  }, []);

  const onChangeHandle = (name: string, value: any) => {
    setConfig({
      ...config,
      [name]: value,
    });
  };

  const handleUpdate = (values: any) => {
    const updatePromises = Object.keys(values).map((key) => {
      return updateConfig({ key, value: values[key] });
    });

    Promise.all(updatePromises)
      .then(() => {
        message.success('配置更新成功');
        refreshFrontendCache().then((res) => {
          if (res) {
            message.success('刷新缓存成功');
          } else {
            message.error('刷新缓存失败');
          }
        });
      })
      .catch(() => {
        message.error('配置更新失败');
      });
  };

  return (
    <PageContainer title={'网站配置'}>
      <Form onFinish={handleUpdate} ref={formRef} layout="vertical">
        <Form.Item name="AUTHOR_AVATAR" label="作者头像">
          <Input
            onChange={(e) => {
              onChangeHandle('AUTHOR_AVATAR', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="AUTHOR_GITHUB" label="作者 GitHub">
          <Input
            onChange={(e) => {
              onChangeHandle('AUTHOR_GITHUB', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="AUTHOR_HOME" label="作者主页">
          <Input
            onChange={(e) => {
              onChangeHandle('AUTHOR_HOME', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="AUTHOR_INFO" label="作者简介">
          <Input
            onChange={(e) => {
              onChangeHandle('AUTHOR_INFO', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="AUTHOR_NAME" label="作者姓名">
          <Input
            onChange={(e) => {
              onChangeHandle('AUTHOR_NAME', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="AUTHOR_WELCOME" label="欢迎信息">
          <Input
            onChange={(e) => {
              onChangeHandle('AUTHOR_WELCOME', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="HOME_GITHUB" label="主页 GitHub">
          <Input
            onChange={(e) => {
              onChangeHandle('HOME_GITHUB', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="HOME_SLOGAN" label="主页标语">
          <Input
            onChange={(e) => {
              onChangeHandle('HOME_SLOGAN', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="HOME_SLOGAN_EN" label="主页标语（英文）">
          <Input
            onChange={(e) => {
              onChangeHandle('HOME_SLOGAN_EN', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="HOME_TITLE" label="主页标题">
          <Input
            onChange={(e) => {
              onChangeHandle('HOME_TITLE', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="WEBSITE_AUTHOR" label="网站作者">
          <Input
            onChange={(e) => {
              onChangeHandle('WEBSITE_AUTHOR', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="WEBSITE_COPYRIGHT" label="网站版权">
          <Input
            onChange={(e) => {
              onChangeHandle('WEBSITE_COPYRIGHT', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="WEBSITE_CREATE_TIME" label="网站创建时间">
          <Input
            onChange={(e) => {
              onChangeHandle('WEBSITE_CREATE_TIME', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="WEBSITE_DESCRIPTION" label="网站描述">
          <Input
            onChange={(e) => {
              onChangeHandle('WEBSITE_DESCRIPTION', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="WEBSITE_FAVICON" label="网站图标">
          <Input
            onChange={(e) => {
              onChangeHandle('WEBSITE_FAVICON', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="WEBSITE_ICP" label="ICP 备案号">
          <Input
            onChange={(e) => {
              onChangeHandle('WEBSITE_ICP', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="WEBSITE_KEYWORDS" label="网站关键词">
          <Input
            onChange={(e) => {
              onChangeHandle('WEBSITE_KEYWORDS', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="WEBSITE_LOGO" label="网站 Logo">
          <Input
            onChange={(e) => {
              onChangeHandle('WEBSITE_LOGO', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="WEBSITE_MPS" label="MPS 备案号">
          <Input
            onChange={(e) => {
              onChangeHandle('WEBSITE_MPS', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="WEBSITE_NAME" label="网站名称">
          <Input
            onChange={(e) => {
              onChangeHandle('WEBSITE_NAME', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item name="WEBSITE_URL" label="网站 URL">
          <Input
            onChange={(e) => {
              onChangeHandle('WEBSITE_URL', e.target.value);
            }}
          />
        </Form.Item>
        <Form.Item>
          <Button type="primary" htmlType="submit">
            更新配置
          </Button>
        </Form.Item>
      </Form>
    </PageContainer>
  );
};

export default ConfigPage;
