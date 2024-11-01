'use client';

import React, { useRef } from 'react';
import { motion, useInView } from 'framer-motion';
import RecentArticleItem from '@/app/home/recent/RecentArticleItem';

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
  hidden: { opacity: 0, x: -20 },
  visible: {
    opacity: 1,
    x: 0,
    transition: {
      duration: 1,
    },
  },
};

type Article = {
  id: string;
  title: string;
  createdAt: string;
};

interface RecentArticleMotionProps {
  list: Article[];
}

export default function RecentArticleMotion({ list }: RecentArticleMotionProps) {
  const ref = useRef<HTMLDivElement>(null);
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-expect-error
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
