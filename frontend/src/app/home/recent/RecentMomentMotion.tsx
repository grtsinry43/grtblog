'use client';

import React, { useRef } from 'react';
import { motion, useInView } from 'framer-motion';
import RecentMomentItem from '@/app/home/recent/RecentMomentItem';
import { StatusUpdate } from '@/types';

const staggerContainer = {
  hidden: { opacity: 1 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1,
    },
  },
};

const itemMotion = {
  hidden: { 
    opacity: 0, 
    y: 8,
  },
  visible: {
    opacity: 1,
    y: 0,
    transition: {
      duration: 0.4,
      ease: "easeOut",
    },
  },
};

interface RecentMomentMotionProps {
  list: StatusUpdate[];
}

export default function RecentMomentMotion({ list }: RecentMomentMotionProps) {
  const ref = useRef<HTMLDivElement>(null);
  
  const inView = useInView(ref, {
    once: true,
    margin: "-20px",
  });

  return (
    <motion.div
      ref={ref}
      variants={staggerContainer}
      initial="hidden"
      animate={inView ? 'visible' : 'hidden'}
    >
      <div className="space-y-0">
        {list.map((item: StatusUpdate, index: number) => (
          <motion.div 
            key={`${item.id}-${index}`} 
            variants={itemMotion}
          >
            <RecentMomentItem statusUpdate={item} />
          </motion.div>
        ))}
      </div>
    </motion.div>
  );
}
