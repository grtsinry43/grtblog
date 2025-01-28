import { history } from '@@/exports';
import { Modal } from 'antd';
import { useEffect } from 'react';

/**
 * 路由离开确认钩子（支持自定义弹窗）
 * @param {string} message 提示消息
 * @param {string} title 弹窗标题
 */
const useRouteLeaveConfirm = (
  message: string = '确定要离开当前页面吗？请确认数据已保存。',
  title: string = '提示',
) => {
  useEffect(() => {
    const unblock = history.block((transition) => {
      const { location, action } = transition;

      Modal.confirm({
        title,
        content: message,
        okText: '好！',
        cancelText: '不要啦，我再看看',
        onOk: () => {
          unblock(); // 取消拦截
          if (action === 'PUSH') {
            history.push(location.pathname);
          } else if (action === 'REPLACE') {
            history.replace(location.pathname);
          } else {
            history.go(1);
          }
        },
      });

      // 阻止默认跳转行为
      return false;
    });

    return () => {
      unblock();
    };
  }, [message, title]); // message 或 title 变化时重新注册拦截器
};

export default useRouteLeaveConfirm;
