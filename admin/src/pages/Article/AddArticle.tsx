import ArticleForm from '@/pages/Article/ArticleForm';
import ArticleController from '@/services/article/ArticleController';
import { AddArticleApiParams } from '@/services/article/typings';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { message } from 'antd';
import { useState } from 'react';
import {refreshFrontendCache} from "@/services/refersh";
import useRouteLeaveConfirm from "@/hooks/use-route-leave-confirm";

const AddArticle = () => {
  const navigate = useNavigate();
  useRouteLeaveConfirm();
  const [newArticleInfo, setNewArticleInfo] = useState<AddArticleApiParams>({
    title: '',
    content: '',
    cover: '',
    shortUrl: '',
    categoryId: '',
    isPublished: false,
  });

  const submitHandle = (content: string) => {
    if (!content) {
      message.error('文章内容不能为空');
      return;
    }
    ArticleController.addArticle({
      ...newArticleInfo,
      content,
    }).then((res) => {
      if (res) {
        message.success('文章添加成功');
        refreshFrontendCache().then((res) => {
          if (res) {
            message.success('刷新缓存成功');
          } else {
            message.error('刷新缓存失败');
          }
        });
        navigate('/article/list');
      }
    });
  };

  return (
    <PageContainer title={'添加文章'}>
      <ArticleForm
        type={'add'}
        articleInfo={newArticleInfo}
        setArticleInfo={setNewArticleInfo}
        submitHandle={submitHandle}
      />
    </PageContainer>
  );
};

export default AddArticle;
