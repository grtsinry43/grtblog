import useRouteLeaveConfirm from '@/hooks/use-route-leave-confirm';
import ArticleForm from '@/pages/Article/ArticleForm';
import ArticleController from '@/services/article/ArticleController';
import { refreshFrontendCache } from '@/services/refersh';
import { useParams } from '@@/exports';
import { PageContainer } from '@ant-design/pro-components';
import { useNavigate } from '@umijs/max';
import { message } from 'antd';
import { useEffect, useState } from 'react';

const EditArticle = () => {
  // 获取传递过来的 id
  const { id } = useParams<{ id: string }>(); // 获取可能传递过来的 id

  useRouteLeaveConfirm();

  const navigate = useNavigate();

  const [articleInfo, setArticleInfo] = useState<any>(null);

  useEffect(() => {
    async function fetchData() {
      if (id) {
        // 根据文章 id 获取该文章具体的信息
        const { data } = await ArticleController.getArticleDetail(id);

        setArticleInfo(data);
      }
    }

    fetchData();
  }, [id]);

  const submitHandle = (content: string) => {
    if (!content) {
      message.error('文章内容不能为空');
      return;
    }
    if (id) {
      ArticleController.editArticle(id, {
        ...articleInfo,
        content,
      }).then((res) => {
        if (res) {
          message.success('文章修改成功');
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
    }
  };

  return (
    <PageContainer title={'编辑文章'}>
      <ArticleForm
        type={'edit'}
        articleInfo={articleInfo}
        setArticleInfo={setArticleInfo}
        submitHandle={submitHandle}
      />
    </PageContainer>
  );
};

export default EditArticle;
