import MomentController from '@/services/moment/MomentController';
import { AddMomentApiParams } from '@/services/moment/typings';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { message } from 'antd';
import { useState } from 'react';
import MomentForm from './MomentForm';
import {refreshFrontendCache} from "@/services/refersh";
import useRouteLeaveConfirm from "@/hooks/use-route-leave-confirm";

const AddMoment = () => {
  const navigate = useNavigate();
  useRouteLeaveConfirm();
  const [newMomentInfo, setNewMomentInfo] = useState<AddMomentApiParams>({
    img: '',
    isOriginal: false,
    isTop: false,
    shortUrl: '',
    title: '',
    content: '',
    cover: '',
    categoryId: '',
    isPublished: false,
  });

  const submitHandle = (content: string) => {
    if (!content) {
      message.error('内容不能为空');
      return;
    }
    MomentController.addMoment({
      ...newMomentInfo,
      content,
    }).then((res) => {
      if (res) {
        message.success('动态添加成功');
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
  };

  return (
    <PageContainer title={'添加动态'}>
      <MomentForm
        type={'add'}
        momentInfo={newMomentInfo}
        setMomentInfo={setNewMomentInfo}
        submitHandle={submitHandle}
      />
    </PageContainer>
  );
};

export default AddMoment;
