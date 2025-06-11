import React from 'react';
import {getRelatedArticles} from '@/api/article';
import {ArticlePreview} from '@/types';
import RecommendCard from '@/components/article/RecommendCard';
import {ScrollArea} from "@radix-ui/themes";

const RelatedRecommend = async ({id}: { id: string }) => {
    const relatedRecommend = await getRelatedArticles(id);
    
    return (
        <div>
            <h1 style={{
                fontSize: '1.5rem',
                fontWeight: 'bold',
                marginBottom: '1rem',
                marginTop: '1rem',
            }}> 相关推荐 </h1>
            <ScrollArea>
                <div className="flex space-x-4">
                    {(relatedRecommend || []).map((recommend: ArticlePreview) => (
                        <RecommendCard key={recommend.id} item={recommend}/>
                    ))}
                </div>
            </ScrollArea>
        </div>
    );
};

export default RelatedRecommend;
