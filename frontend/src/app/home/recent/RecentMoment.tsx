import React from 'react';
import Link from 'next/link';
import { ArrowRightIcon, ChatBubbleIcon } from '@radix-ui/react-icons';
import RecentMomentMotion from '@/app/home/recent/RecentMomentMotion';
import { StatusUpdate } from '@/types';

interface RecentMomentProps {
  shareList: StatusUpdate[];
}

const RecentMoment = ({ shareList }: RecentMomentProps) => {
  return (
    <div className="relative flex-1 flex flex-col min-h-[600px]">
      {/* 背景装饰层 */}
      <div className="absolute inset-0 bg-gradient-to-bl from-foreground/[0.02] via-transparent to-foreground/[0.01] rounded-2xl pointer-events-none" />
      <div className="absolute top-4 left-6 opacity-5 pointer-events-none">
        <ChatBubbleIcon className="w-16 h-16" />
      </div>
      
      {/* 主标题 */}
      <div className="relative flex items-center gap-3 text-2xl font-bold text-start mb-8">
        <div className="flex-1 h-px bg-gradient-to-l from-foreground/10 via-foreground/5 to-transparent mt-1"></div>
        <span>随手记录</span>
        <div className="w-1 h-8 bg-gradient-to-b from-primary/60 to-primary/20 rounded-full"></div>
      </div>
      
      {/* 随手记录列表容器 */}
      <div className="relative flex-1">
        <RecentMomentMotion list={shareList} />
      </div>
      
      {/* 查看更多按钮 */}
      <div className="relative text-end mt-auto pt-8 group">
        <Link 
          href="/moments" 
          className="inline-flex items-center gap-1 px-3 py-1.5 rounded
                     hover:bg-foreground/5 transition-all duration-200
                     text-sm opacity-60 hover:opacity-100
                     relative overflow-hidden"
        >
          <div className="absolute inset-0 bg-gradient-to-r from-transparent via-foreground/5 to-transparent 
                          translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-500"></div>
          <span className="relative">查看更多</span>
          <ArrowRightIcon className="w-4 h-4 transition-transform duration-200 group-hover:translate-x-0.5 relative" />
        </Link>
      </div>
    </div>
  );
};

export default RecentMoment;
