import useRouteLeaveConfirm from '@/hooks/use-route-leave-confirm';
import PageForm from '@/pages/Page/PageForm';
import PageController from '@/services/page/PageController';
import { AddPageApiParams } from '@/services/page/typings';
import { refreshFrontendCache } from '@/services/refersh';
import { useNavigate } from '@@/exports';
import { PageContainer } from '@ant-design/pro-components';
import { message } from 'antd';
import { useState } from 'react';

const AddPage = () => {
  const navigate = useNavigate();
  const [newPageInfo, setNewPageInfo] = useState<AddPageApiParams>({
    canComment: true,
    content: '',
    description: '',
    enable: false,
    refPath: '',
    title: '',
  });
  useRouteLeaveConfirm();

  const submitHandle = (content: string) => {
    if (!content) {
      message.error('内容不能为空');
      return;
    }
    PageController.addPage({
      ...newPageInfo,
      content,
    }).then((res) => {
      if (res) {
        message.success('页面添加成功');
        refreshFrontendCache().then((res) => {
          if (res) {
            message.success('刷新缓存成功');
          } else {
            message.error('刷新缓存失败');
          }
        });
        navigate('/pages/list');
      }
    });
  };

  return (
    <PageContainer title={'添加页面'}>
      <PageForm
        type={'add'}
        pageInfo={newPageInfo}
        setPageInfo={setNewPageInfo}
        submitHandle={submitHandle}
      />
    </PageContainer>
  );
};

export default AddPage;
