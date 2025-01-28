import useRouteLeaveConfirm from '@/hooks/use-route-leave-confirm';
import PageForm from '@/pages/Page/PageForm';
import PageController from '@/services/page/PageController';
import { refreshFrontendCache } from '@/services/refersh';
import { useNavigate, useParams } from '@@/exports';
import { PageContainer } from '@ant-design/pro-components';
import { message } from 'antd';
import { useEffect, useState } from 'react';

const EditPage = () => {
  const { id } = useParams<{ id: string }>();
  useRouteLeaveConfirm();

  const navigate = useNavigate();

  const [pageInfo, setPageInfo] = useState<any>(null);

  useEffect(() => {
    async function fetchData() {
      if (id) {
        const { data } = await PageController.getPageDetail(id);
        setPageInfo(data);
      }
    }

    fetchData();
  }, [id]);

  const submitHandle = (content: string) => {
    if (!content) {
      message.error('内容不能为空');
      return;
    }
    if (id) {
      PageController.editPage(id, {
        ...pageInfo,
        content,
      }).then((res) => {
        if (res) {
          message.success('页面成功更新');
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
    }
  };

  return (
    <PageContainer title={'更新页面'}>
      <PageForm
        type={'edit'}
        pageInfo={pageInfo}
        setPageInfo={setPageInfo}
        submitHandle={submitHandle}
      />
    </PageContainer>
  );
};

export default EditPage;
