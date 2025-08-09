import React from 'react';
import CommentForm from '@/components/comment/CommentForm';
import CommentList from '@/components/comment/CommentList';
import {ScrollArea} from "@radix-ui/themes";
import {clsx} from "clsx";
import {article_font} from "@/app/fonts/font";
import InfoCard from "@/components/InfoCard";

// 最后还是用了客户端组件...
const CommentArea = ({id, isModal}: { id: string, isModal?: boolean }) => {
    return (
        <div className={"flex flex-row"} style={{
            marginTop: '3rem',
            height: 'fit-content',
        }}>
            <div style={{
                flex: 1,
            }}>
                <span className={clsx(article_font.className, "text-sm opacity-10")}> COMMENT {id}</span>
                <div className={clsx(article_font.className, "flex")}>
                    <h2 className="text-2xl font-bold"> 发表评论 </h2>
                    <span className="ml-4 self-end text-sm opacity-60"> 来这里畅所欲言吧！</span>
                </div>
                <CommentForm id={id}/>
                {isModal ? (
                    <div style={{
                        paddingLeft: '1rem',
                        paddingRight: '1rem',
                    }}>
                        <CommentList id={id} isModal={isModal}/>
                    </div>
                ) : (
                    <ScrollArea style={{
                        transition: 'all 0.3s',
                        paddingLeft: '1rem',
                        paddingRight: '1rem',
                        height: 'fit-content',
                        maxHeight: 'calc(100vh - 13rem)',
                    }}>
                        <CommentList id={id} isModal={isModal}/>
                    </ScrollArea>
                )}
            </div>
            {!isModal && <div style={{
                height: '100%',
                paddingTop: '3rem',
            }}>
                <InfoCard/>
            </div>
            }
        </div>
    );
};

export default CommentArea;
