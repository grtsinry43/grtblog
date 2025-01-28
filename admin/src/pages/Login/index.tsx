import React, { useState } from 'react';
// @ts-ignore
import ReactCanvasNest from 'react-canvas-nest';
import styles from './Index.module.css';
import UserController from '@/services/user/UserController';
import { Button, Checkbox, Form, Input, message } from 'antd';

const Login: React.FC = () => {
  const [form, setForm] = useState({
    userEmail: '',
    password: '',
    captcha: '',
    remember: true,
  });
  const [captchaUrl, setCaptchaUrl] = useState('/captcha');

  const onFormChange = (key: string, value: any) => {
    setForm({
      ...form,
      [key]: value,
    });
  };

  const refreshCaptcha = () => {
    setCaptchaUrl(`/captcha?${Date.now()}`);
  };

  const onFinish = () => {
    UserController.userLoginApi(form).then((res) => {
      if (res.data) {
        message.success('登录成功，即将跳转到首页');
        location.href = '/';
      } else {
        message.error(res.msg);
        refreshCaptcha();
      }
    });
  };

  return (
      <div className={styles.pageContainer}>
        <ReactCanvasNest
            config={{
              pointColor: '255, 255, 255',
              count: 66,
              follow: true,
            }}
            style={{
              zIndex: 1,
              opacity: 0.5,
              width: '100%',
              height: '100%',
            }}
        />
        <div className={styles.formContainer}>
          <div className={styles.leftContainer}>
            <div className={styles.title}>Grtblog</div>
            <div className={styles.subTitle}>
              今天，又要书写什么故事呢？
            </div>
          </div>
          <Form onFinish={onFinish} className={styles.rightContainer}>
            <div className={styles.rightTitle}>欢迎登录</div>
            <div className={styles.rightSubTitle}>Welcome to Grtblog</div>
            <Form.Item
                name="userEmail"
                rules={[{ required: true, message: '请输入用户名' }]}
                className={styles.inputContainer}
            >
              <Input
                  type="text"
                  placeholder="请输入用户名"
                  onChange={(e) => onFormChange('userEmail', e.target.value)}
                  className={styles.input}
              />
            </Form.Item>
            <Form.Item
                name="password"
                rules={[{ required: true, message: '请输入密码' }]}
                className={styles.inputContainer}
            >
              <Input
                  type="password"
                  placeholder="请输入密码"
                  onChange={(e) => onFormChange('password', e.target.value)}
                  className={styles.input}
              />
            </Form.Item>
            <Form.Item
                name="captcha"
                rules={[{ required: true, message: '请输入验证码' }]}
                className={styles.inputContainer}
            >
              <div className={styles.captchaContainer}>
                <Input
                    type="text"
                    placeholder="请输入验证码"
                    onChange={(e) => onFormChange('captcha', e.target.value)}
                    className={styles.input}
                />
                <img
                    src={captchaUrl || "/placeholder.svg"}
                    alt="captcha"
                    className={styles.captcha}
                    onClick={refreshCaptcha}
                />
              </div>
            </Form.Item>
            <div className={styles.inputContainer}>
              <Checkbox
                  onChange={(e) => {
                    onFormChange('remember', e.target.checked);
                  }}
                  className={styles.remember}
              />
              <span className={styles.rememberText}>7 天内保持登录状态</span>
            </div>
            <Form.Item className={styles.inputContainer}>
              <Button
                  htmlType={'submit'}
                  type={'primary'}
                  className={styles.loginButton}
              >
                登录
              </Button>
            </Form.Item>
          </Form>
        </div>
      </div>
  );
};

export default Login;
