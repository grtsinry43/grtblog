import MomentController from '@/services/moment/MomentController';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate, useParams } from '@umijs/max';
import { message } from 'antd';
import { useEffect, useState } from 'react';
import MomentForm from './MomentForm';
import {refreshFrontendCache} from "@/services/refersh";
import useRouteLeaveConfirm from "@/hooks/use-route-leave-confirm";

const EditMoment = () => {
  const navigate = useNavigate();
  const { id } = useParams<string>();
  const [momentInfo, setMomentInfo] = useState<any>(null);
  useRouteLeaveConfirm();

  useEffect(() => {
    async function fetchData() {
      if (id) {
        // 根据文章 id 获取该文章具体的信息
        const { data } = await MomentController.getMomentDetail(id);

        setMomentInfo(data);
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
      MomentController.editMoment(id, {
        ...momentInfo,
        content,
      }).then((res) => {
        if (res) {
          message.success('动态修改成功');
          refreshFrontendCache().then((res) => {
            if (res) {
              message.success('刷新缓存成功');
            } else {
              message.error('刷新缓存失败');
            }
          });
          navigate('/moment/list');
        }
      });
    }
  };

  return (
    <PageContainer title={'编辑动态'}>
      <MomentForm
        type={'edit'}
        momentInfo={momentInfo}
        setMomentInfo={setMomentInfo}
        submitHandle={submitHandle}
      />
    </PageContainer>
  );
};

export default EditMoment;
