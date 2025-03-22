'use client';

import React, { useRef } from 'react';
import { motion, useInView } from 'framer-motion';
import RecentArticleItem from '@/app/home/recent/RecentArticleItem';
import { Article } from '@/types';

const staggerContainer = {
  hidden: { opacity: 1 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.2,
    },
  },
};

const itemMotion = {
  hidden: { opacity: 0, x: -20, filter: 'blur(10px)', boxShadow: '0 0 10px rgba(var(--foreground), 0.5)' },
  visible: {
    opacity: 1,
    filter: 'blur(0px)',
    boxShadow: 'none',
    x: 0,
    transition: {
      duration: 1,
    },
  },
};

interface RecentArticleMotionProps {
  list: Article[];
}

export default function RecentArticleMotion({ list }: RecentArticleMotionProps) {
  const ref = useRef<HTMLDivElement>(null);
  
  
  const inView = useInView(ref, {
    once: true,
  });

  return (
    <motion.div
      ref={ref}
      variants={staggerContainer}
      initial="hidden"
      animate={inView ? 'visible' : 'hidden'}
    >
      <div className="relative">
        {list.map((item: Article, index: number) => (
          <motion.div key={index} variants={itemMotion}>
            <div className="relative">
              <RecentArticleItem article={item} />
            </div>
          </motion.div>
        ))}
      </div>
    </motion.div>
  );
}
