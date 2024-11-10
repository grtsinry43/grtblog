import React from 'react';
import { getRelatedArticles } from '@/api/article';
import { ArticlePreview } from '@/types';
import RecommendCard from '@/components/article/RecommendCard';

const RelatedRecommend = async ({ id }: { id: string }) => {
  const relatedRecommend = await getRelatedArticles(id);
  console.log(relatedRecommend);
  return (
    <div>
      <h1>id 推荐区域 </h1>
      <div
        style={{
          scrollBehavior: 'smooth',
          overflowX: 'scroll',
          display: 'flex',
          flexDirection: 'row',
        }}
        className="flex flex-row space-x-4"
      >
        {relatedRecommend.map((recommend: ArticlePreview) => (
          <RecommendCard key={recommend.id} item={recommend} />
        ))}
      </div>
    </div>
  );
};

export default RelatedRecommend;