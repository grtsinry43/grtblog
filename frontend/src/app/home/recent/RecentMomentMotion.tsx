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
      staggerChildren: 0.2,
    },
  },
};

const itemMotion = {
  hidden: {
    opacity: 0,
    y: 20,
    rotateX: -90,
    filter: 'blur(10px)',
    boxShadow: '0 0 10px rgba(var(--foreground), 0.5)',
  },
  visible: {
    opacity: 1,
    y: 0,
    rotateX: 0,
    filter: 'blur(0px)',
    border: 'none',
    boxShadow: 'none',
    transition: {
      duration: 1,
    },
  },
};

interface RecentMomentMomentProps {
  list: StatusUpdate[];
}

const RecentMomentMotion = ({ list }: RecentMomentMomentProps) => {
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
      <div>
        {list.map((item: StatusUpdate, index: number) => (
          <motion.div key={index} variants={itemMotion}>
            <div className="relative">
              <RecentMomentItem key={item.id} statusUpdate={item} />
            </div>
          </motion.div>
        ))}
      </div>
    </motion.div>
  );
};

export default RecentMomentMotion;
